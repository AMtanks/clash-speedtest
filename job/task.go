package job

import (
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/starudream/go-lib/core/v2/config"
	"github.com/starudream/go-lib/core/v2/slog"
	"github.com/starudream/go-lib/core/v2/utils/fmtutil"
	"github.com/starudream/go-lib/core/v2/utils/maputil"
	"github.com/starudream/go-lib/core/v2/utils/signalutil"
	"github.com/starudream/go-lib/tablew/v2"

	"github.com/starudream/clash-speedtest/api/clash"
	"github.com/starudream/clash-speedtest/util"
)

type Task struct {
	ClashAddr   string `yaml:"clash.addr"`
	ClashSecret string `yaml:"clash.secret"`
	ClashProxy  string `yaml:"clash.proxy"`

	Size     int      `yaml:"size"`
	Threads  int      `yaml:"threads"`
	Download string   `yaml:"download"`
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes"`
	Confirm  bool     `yaml:"confirm"`
	Output   string   `yaml:"output"`
	
	// Ping configuration
	EnablePing   bool `yaml:"ping"`
	PingInterval int  `yaml:"ping-interval"`
	PingTimeout  int  `yaml:"ping-timeout"`
	
	// Output configuration
	OutputFormat string `yaml:"format"` // txt or png
	
	// Concurrent nodes configuration
	ConcurrentNodes int `yaml:"concurrent"` // number of nodes to test concurrently

	clash   *clash.Client
	version *clash.Version
	config  *clash.Config
	proxies []*clash.Proxy

	results maputil.SyncMap[string, *Result]
	
	pingStop chan struct{} // signal to stop ping goroutines
}

func Run() error {
	t := &Task{}
	err := config.Unmarshal("", t)
	if err != nil {
		return err
	}

	if t.Threads < 1 || t.Threads > 16 {
		t.Threads = 2
	}
	
	// Set default concurrent nodes
	if t.ConcurrentNodes < 1 {
		t.ConcurrentNodes = 1 // default serial testing
	}
	if t.ConcurrentNodes > 5 {
		t.ConcurrentNodes = 5 // limit max concurrent to avoid Clash overload
	}
	
	// Set default ping values
	if t.PingInterval <= 0 {
		t.PingInterval = 60 // default 60 seconds
	}
	if t.PingTimeout <= 0 {
		t.PingTimeout = 5000 // default 5000 ms
	}
	
	// Set default output format
	if t.OutputFormat == "" {
		t.OutputFormat = "txt"
	}

	err = t.Clash()
	if err != nil {
		return err
	}

	if len(t.proxies) == 0 {
		return fmt.Errorf("no proxies found")
	}

	slog.Info("clash version: %s", t.version.Version,
		slog.Bool("premium", t.version.Premium),
		slog.Bool("meta", t.version.Meta),
		slog.String("mode", string(t.config.Mode)),
		slog.String("proxy", t.ClashProxy),
		slog.String("includes", strings.Join(t.Includes, ",")),
		slog.String("excludes", strings.Join(t.Excludes, ",")),
		slog.Int("total", len(t.proxies)),
		slog.Int("threads", t.Threads),
		slog.Int("concurrent", t.ConcurrentNodes),
	)

	fmt.Print(tablew.Structs(t.proxies))

	if !t.Confirm {
		input := fmtutil.Scan("confirm to start? [y/n]: ")
		if !strings.EqualFold(input, "y") {
			fmt.Println("cancel by user")
			return nil
		}
	}

	err = t.clash.SetMode(clash.ModeGlobal)
	if err != nil {
		return err
	}

	reset := func() {
		err = t.clash.SetMode(t.config.Mode)
		if err != nil {
			slog.Error("set mode error: %v", err)
		}
	}
	go func() { signalutil.Defer(reset).Done() }()
	defer func() { reset() }()

	defer func(start time.Time) {
		slog.Info("took %s", time.Since(start).Truncate(time.Millisecond))
	}(time.Now())
	
	// Start ping monitoring if enabled
	if t.EnablePing {
		t.pingStop = make(chan struct{})
		go t.pingMonitor()
		slog.Info("ping monitoring enabled", 
			slog.Int("interval", t.PingInterval),
			slog.Int("timeout", t.PingTimeout))
	}
	
	// Ensure ping stops and then render
	defer func() {
		if t.EnablePing && t.pingStop != nil {
			close(t.pingStop)
			time.Sleep(100 * time.Millisecond) // wait for ping goroutine to exit
		}
		t.Render()
	}()

	// Use worker pool for concurrent node testing
	type proxyJob struct {
		proxy *clash.Proxy
		index int
	}
	
	proxyQueue := make(chan proxyJob, len(t.proxies))
	var wg sync.WaitGroup
	
	// Start concurrent workers
	for w := 0; w < t.ConcurrentNodes; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range proxyQueue {
				proxy := job.proxy
				index := fmt.Sprintf("%d/%d", job.index+1, len(t.proxies))
				
			slog.Info("worker %d start proxy test: %s", workerID+1, proxy.Name, 
				slog.String("type", proxy.Type), 
				slog.String("index", index))
			
			// 测速失败时最多重试2次
			var result *Result
			var err2 error
			maxRetries := 2
			
			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					slog.Info("worker %d retry %d/%d for proxy: %s", workerID+1, attempt, maxRetries, proxy.Name)
				}
				
				result, err2 = t.Test(proxy)
				if err2 == nil {
					break // 成功，跳出重试循环
				}
				
				if attempt < maxRetries {
					slog.Warn("worker %d test failed (attempt %d/%d): %v", workerID+1, attempt+1, maxRetries+1, err2,
						slog.String("proxy", proxy.Name))
				} else {
					slog.Error("worker %d test failed after %d attempts: %v", workerID+1, maxRetries+1, err2, 
						slog.String("proxy", proxy.Name),
						slog.String("index", index))
				}
			}
			
			if err2 != nil {
				continue // 所有重试都失败，跳过此节点
			}
			
			t.results.Store(proxy.Name, result)
				
				// Perform initial ping test immediately after speedtest
				if t.EnablePing {
					delay, err3 := t.clash.TestProxyDelay(proxy.Name, uint16(t.PingTimeout), "")
					if err3 != nil {
						slog.Debug("worker %d initial ping %s failed: %v", workerID+1, proxy.Name, err3)
					} else {
						result.AddPing(delay)
						slog.Debug("worker %d initial ping %s: %dms", workerID+1, proxy.Name, delay)
					}
				}
				
				slog.Info("worker %d proxy test done: %s", workerID+1, proxy.Name, slog.String("index", index))
			}
		}(w)
	}
	
	// Add all proxies to the queue
	for i, proxy := range t.proxies {
		proxyQueue <- proxyJob{proxy: proxy, index: i}
	}
	close(proxyQueue)
	
	// Wait for all workers to complete
	wg.Wait()

	return nil
}

func (t *Task) Clash() error {
	u, err := url.Parse(t.ClashAddr)
	if err != nil {
		return err
	}
	h, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	t.clash = clash.NewClient(t.ClashAddr, t.ClashSecret)

	t.version, err = t.clash.GetVersion()
	if err != nil {
		return err
	}

	t.config, err = t.clash.GetConfig()
	if err != nil {
		return err
	}

	if t.ClashProxy == "" {
		t.ClashProxy, err = t.config.Proxy(h)
		if err != nil {
			return err
		}
	}

	providers, err := t.clash.GetProviderProxies()
	if err != nil {
		return err
	}

	for _, proxy := range providers.FilterProxies() {
		if len(t.Includes) > 0 && !util.Contains(t.Includes, proxy.Name) {
			continue
		}
		if len(t.Excludes) > 0 && util.Contains(t.Excludes, proxy.Name) {
			continue
		}
		t.proxies = append(t.proxies, proxy)
	}

	return nil
}

func (t *Task) Render() {
	fi, err := os.Stat(t.Output)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Error("stat error: %v", err)
			return
		}
		err = os.MkdirAll(t.Output, 0755)
		if err != nil {
			slog.Error("mkdir error: %v", err)
			return
		}
	} else if !fi.IsDir() {
		slog.Error("output is not a directory")
		return
	}

	// Determine output format
	if t.OutputFormat == "png" || t.OutputFormat == "image" {
		t.RenderImage()
	} else {
		t.RenderText()
	}
}

func (t *Task) RenderText() {
	filename := filepath.Join(t.Output, fmt.Sprintf("%s.txt", time.Now().Format("20060102-150405")))

	table := tablew.Render(func(w *tablew.Table) {
		w.SetAlignment(tablew.ALIGN_CENTER)

	var headers []string
	if t.EnablePing {
		headers = []string{"name", "type", "avg-speed", "ping-min", "ping-max", "ping-avg"}
	} else {
		headers = []string{"name", "type", "avg-speed"}
	}
	w.SetHeader(headers)
		
		for i := 0; i < len(t.proxies); i++ {
			proxy := t.proxies[i]
			res, exists := t.results.Load(proxy.Name)
			if !exists {
				continue
			}
			
		avgSpeed := res.GetAvgSpeed()
		
		var avgSpeedStr string
		if avgSpeed == 0 {
			avgSpeedStr = "无法连接"
		} else {
			avgSpeedStr = fmt.Sprintf("%.2f MB/s", avgSpeed)
		}

		var row []string
		var colors []tablew.Colors

		if t.EnablePing {
			pingMin := res.pingMin
			pingMax := res.pingMax
			pingAvg := res.GetPingAvg()
			
			var pingMinStr, pingMaxStr, pingAvgStr string
			if res.pingCount == 0 {
				pingMinStr = "N/A"
				pingMaxStr = "N/A"
				pingAvgStr = "N/A"
			} else {
				pingMinStr = fmt.Sprintf("%dms", pingMin)
				pingMaxStr = fmt.Sprintf("%dms", pingMax)
				pingAvgStr = fmt.Sprintf("%dms", pingAvg)
			}

			row = []string{
				proxy.Name,
				proxy.Type,
				avgSpeedStr,
				pingMinStr,
				pingMaxStr,
				pingAvgStr,
			}

			colors = []tablew.Colors{
				{tablew.Bold},                       // 名称
				{},                                  // 类型
				{tablew.Bold, speedColor(avgSpeed)}, // 平均速度
				{tablew.Bold, pingColor(pingMin)},   // Ping 最小
				{tablew.Bold, pingColor(pingMax)},   // Ping 最大
				{tablew.Bold, pingColor(pingAvg)},   // Ping 平均
			}
		} else {
			row = []string{
				proxy.Name,
				proxy.Type,
				avgSpeedStr,
			}

			colors = []tablew.Colors{
				{tablew.Bold},                       // 名称
				{},                                  // 类型
				{tablew.Bold, speedColor(avgSpeed)}, // 平均速度
			}
		}

			w.Rich(row, colors)
		}
	})
	fmt.Print(table)

	err := os.WriteFile(filename, []byte(table), 0644)
	if err != nil {
		slog.Error("write file error: %v", err)
		return
	}
}

func connColor(d time.Duration) int {
	if d >= time.Second {
		return tablew.FgRedColor
	}
	return 0
}

func speedColor(speed float64) int {
	if speed >= 5.0 {
		return tablew.FgGreenColor
	} else if speed >= 2.0 {
		return tablew.FgYellowColor
	}
	return tablew.FgRedColor
}

func pingColor(p uint16) int {
	if p == 0 {
		return 0
	}
	if p < 100 {
		return tablew.FgGreenColor
	} else if p < 300 {
		return tablew.FgYellowColor
	}
	return tablew.FgRedColor
}

// pingMonitor performs periodic ping tests on all proxies
func (t *Task) pingMonitor() {
	ticker := time.NewTicker(time.Duration(t.PingInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-t.pingStop:
			return
		case <-ticker.C:
			t.pingAllProxies()
		}
	}
}

// pingAllProxies pings all proxies once
func (t *Task) pingAllProxies() {
	t.results.Range(func(name string, result *Result) bool {
		delay, err := t.clash.TestProxyDelay(name, uint16(t.PingTimeout), "")
		if err != nil {
			slog.Debug("periodic ping %s failed: %v", name, err)
			return true
		}
		result.AddPing(delay)
		slog.Debug("periodic ping %s: %dms", name, delay)
		return true
	})
}

// RenderImage renders results as PNG image
func (t *Task) RenderImage() {
	filename := filepath.Join(t.Output, fmt.Sprintf("%s.png", time.Now().Format("20060102-150405")))

	var headers []string
	if t.EnablePing {
		headers = []string{"节点名称", "类型", "平均速度", "Ping最小", "Ping最大", "Ping平均"}
	} else {
		headers = []string{"节点名称", "类型", "平均速度"}
	}

	var rows [][]string
	var colors [][]color.Color

	for i := 0; i < len(t.proxies); i++ {
		proxy := t.proxies[i]
		res, exists := t.results.Load(proxy.Name)
		if !exists {
			continue
		}

		avgSpeed := res.GetAvgSpeed()
		
		var avgSpeedStr string
		if avgSpeed == 0 {
			avgSpeedStr = "无法连接"
		} else {
			avgSpeedStr = fmt.Sprintf("%.2f MB/s", avgSpeed)
		}

		var row []string
		var rowColors []color.Color

		if t.EnablePing {
			pingMin := res.pingMin
			pingMax := res.pingMax
			pingAvg := res.GetPingAvg()
			
			var pingMinStr, pingMaxStr, pingAvgStr string
			if res.pingCount == 0 {
				pingMinStr = "N/A"
				pingMaxStr = "N/A"
				pingAvgStr = "N/A"
			} else {
				pingMinStr = fmt.Sprintf("%dms", pingMin)
				pingMaxStr = fmt.Sprintf("%dms", pingMax)
				pingAvgStr = fmt.Sprintf("%dms", pingAvg)
			}

			row = []string{
				proxy.Name,
				proxy.Type,
				avgSpeedStr,
				pingMinStr,
				pingMaxStr,
				pingAvgStr,
			}

			rowColors = []color.Color{
				color.RGBA{44, 62, 80, 255},       // 名称
				color.RGBA{127, 140, 141, 255},    // 类型
				getSpeedColorRGBA(avgSpeed),       // 平均速度
				getPingColorRGBA(pingMin),         // Ping 最小
				getPingColorRGBA(pingMax),         // Ping 最大
				getPingColorRGBA(pingAvg),         // Ping 平均
			}
		} else {
			row = []string{
				proxy.Name,
				proxy.Type,
				avgSpeedStr,
			}

			rowColors = []color.Color{
				color.RGBA{44, 62, 80, 255},       // 名称
				color.RGBA{127, 140, 141, 255},    // 类型
				getSpeedColorRGBA(avgSpeed),       // 平均速度
			}
		}

		rows = append(rows, row)
		colors = append(colors, rowColors)
	}

	// Create image config
	config := util.DefaultTableImageConfig()
	config.Title = "Clash 节点测速结果"
	config.Headers = headers
	config.Rows = rows
	config.Colors = colors

	// Render image
	dc, err := util.RenderTableImage(config)
	if err != nil {
		slog.Error("render image error: %v", err)
		fmt.Printf("⚠️  图片生成失败: %v\n", err)
		fmt.Println("💡 降级使用文本输出...")
		t.RenderText()
		return
	}

	// Save image
	err = dc.SavePNG(filename)
	if err != nil {
		slog.Error("save image error: %v", err)
		fmt.Printf("⚠️  图片保存失败: %v\n", err)
		return
	}

	fmt.Printf("\n✅ 图片已保存到: %s\n", filename)
	slog.Info("image saved: %s", filename)
}

// Helper functions for colors
func getConnTimeColor(d time.Duration) color.Color {
	ms := d.Milliseconds()
	if ms < 100 {
		return color.RGBA{46, 204, 113, 255} // 绿色
	} else if ms < 500 {
		return color.RGBA{241, 196, 15, 255} // 黄色
	}
	return color.RGBA{231, 76, 60, 255} // 红色
}

func getSpeedColorRGBA(speed float64) color.Color {
	if speed >= 5.0 {
		return color.RGBA{46, 204, 113, 255} // 绿色
	} else if speed >= 2.0 {
		return color.RGBA{241, 196, 15, 255} // 黄色
	}
	return color.RGBA{231, 76, 60, 255} // 红色
}

func getPingColorRGBA(ping uint16) color.Color {
	if ping == 0 {
		return color.RGBA{149, 165, 166, 255} // 灰色
	}
	if ping < 100 {
		return color.RGBA{46, 204, 113, 255} // 绿色
	} else if ping < 300 {
		return color.RGBA{241, 196, 15, 255} // 黄色
	}
	return color.RGBA{231, 76, 60, 255} // 红色
}
