package job

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/starudream/go-lib/core/v2/slog"
	"github.com/starudream/go-lib/core/v2/utils/maputil"

	"github.com/starudream/clash-speedtest/api/clash"
	"github.com/starudream/clash-speedtest/api/cloudflare"
	"github.com/starudream/clash-speedtest/api/common"
	"github.com/starudream/clash-speedtest/api/fast"
	"github.com/starudream/clash-speedtest/api/speedtest"
	"github.com/starudream/clash-speedtest/util"
)

type WebTask struct {
	*Task
	broadcast chan<- interface{}
	speedUpdates map[string]*SpeedTracker
	speedMutex sync.RWMutex
	StopChan   <-chan struct{}
	Downloads  []string
}

type SpeedTracker struct {
	lastSpeed float64
	lastUpdate time.Time
}

func NewTaskFromConfig(config map[string]interface{}, broadcast chan<- interface{}) *WebTask {
	// Get download methods (support multiple)
	downloads := getStringSlice(config, "downloads")
	if len(downloads) == 0 {
		downloads = []string{getString(config, "download", "cloudflare")}
	}
	
	// Use first download method as default
	downloadMethod := downloads[0]
	if len(downloads) > 0 {
		downloadMethod = downloads[0]
	}
	
	task := &Task{
		ClashAddr:       getString(config, "clash_addr", "http://127.0.0.1:9090"),
		ClashSecret:     getString(config, "clash_secret", ""),
		ClashProxy:      getString(config, "clash_proxy", ""),
		Size:            getInt(config, "size", 10),
		Threads:         getInt(config, "threads", 1),
		Download:        downloadMethod,
		Includes:        getStringSlice(config, "includes"),
		Excludes:        getStringSlice(config, "excludes"),
		Confirm:         true,
		Output:          getString(config, "output", "output"),
		Timeout:         getInt(config, "timeout", 15),
		EnablePing:      getBool(config, "ping", true),
		PingInterval:    getInt(config, "ping_interval", 20),
		PingTimeout:     getInt(config, "ping_timeout", 5000),
		OutputFormat:    getString(config, "format", "png"),
		ConcurrentNodes: getInt(config, "concurrent", 2),
		results:         maputil.SyncMap[string, *Result]{},
	}
	
	return &WebTask{
		Task:         task,
		broadcast:    broadcast,
		speedUpdates: make(map[string]*SpeedTracker),
		Downloads:    downloads,
	}
}

func (wt *WebTask) Run() error {
	wt.sendLog("info", "Initializing Clash connection...")
	
	err := wt.Clash()
	if err != nil {
		wt.sendLog("error", fmt.Sprintf("Failed to connect to Clash: %v", err))
		return err
	}
	
	if len(wt.proxies) == 0 {
		wt.sendLog("error", "No proxies found")
		return fmt.Errorf("no proxies found")
	}
	
	wt.sendLog("info", fmt.Sprintf("Found %d proxies", len(wt.proxies)))
	wt.sendProxies()
	
	err = wt.clash.SetMode(clash.ModeGlobal)
	if err != nil {
		wt.sendLog("error", fmt.Sprintf("Failed to set global mode: %v", err))
		return err
	}
	
	defer func() {
		err = wt.clash.SetMode(wt.config.Mode)
		if err != nil {
			slog.Error("set mode error: %v", err)
		}
	}()
	
	// Start ping monitoring if enabled - runs throughout entire speedtest
	if wt.EnablePing {
		wt.pingStop = make(chan struct{})
		go wt.pingMonitorWithBroadcast()
		wt.sendLog("info", fmt.Sprintf("Ping monitoring enabled (interval: %ds)", wt.PingInterval))
	}
	
	// Ensure ping monitoring stops before function returns
	defer func() {
		if wt.EnablePing && wt.pingStop != nil {
			close(wt.pingStop)
			// Wait longer to ensure ping goroutine fully exits
			time.Sleep(500 * time.Millisecond)
		}
	}()
	
	// Test with each download method
	downloadMethods := wt.Downloads
	if len(downloadMethods) == 0 {
		downloadMethods = []string{wt.Download}
	}
	
	for _, method := range downloadMethods {
		wt.Download = method
		wt.sendLog("info", fmt.Sprintf("Starting speedtest with %s method...", method))
		
		// Run speedtest with progress updates
		wt.sendLog("info", fmt.Sprintf("Starting speedtest with %d concurrent workers...", wt.ConcurrentNodes))
		err = wt.runSpeedtestWithProgress()
		if err != nil {
			wt.sendLog("error", fmt.Sprintf("Speedtest failed: %v", err))
			return err
		}
	}
	
	wt.sendLog("info", "Speedtest completed, sending results...")
	wt.sendResults()
	
	// Output results to file
	wt.sendLog("info", "Generating output files...")
	wt.Render()
	
	return nil
}

func (wt *WebTask) runSpeedtestWithProgress() error {
	type proxyJob struct {
		proxy *clash.Proxy
		index int
	}
	
	proxyQueue := make(chan proxyJob, len(wt.proxies))
	results := make(chan *Result, len(wt.proxies))
	errors := make(chan error, len(wt.proxies))
	
	// Start workers
	for w := 0; w < wt.ConcurrentNodes; w++ {
		go func(workerID int) {
			for job := range proxyQueue {
				// Check if stop signal received
				select {
				case <-wt.StopChan:
					wt.sendLog("info", fmt.Sprintf("Worker %d received stop signal", workerID+1))
					return
				default:
				}
				
				proxy := job.proxy
				
				wt.sendProgress(job.index+1, len(wt.proxies), proxy.Name, "testing")
				wt.sendLog("info", fmt.Sprintf("Worker %d testing: %s", workerID+1, proxy.Name))
				
				// Create result object early and store it for monitoring
				result := &Result{
					Proxy:     proxy,
					threads:   wt.Threads,
					total:     &common.DownloadResult{},
					downloads: make([]*common.DownloadResult, wt.Threads),
				}
				wt.results.Store(proxy.Name, result)
				
				// Start speed monitoring for this proxy
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(wt.Timeout)*time.Second)
				defer cancel()
				
				go wt.monitorSpeed(ctx, proxy.Name)
				
				testResult, err := wt.TestWithTimeout(proxy, time.Duration(wt.Timeout)*time.Second, result)
				cancel() // Stop speed monitoring
				
				// Update result if test succeeded
				if testResult != nil {
					result = testResult
					wt.results.Store(proxy.Name, result)
				}
				
				isTimeout := false
				if err != nil {
					// Check if it's a timeout error
					if ctx.Err() == context.DeadlineExceeded || strings.Contains(err.Error(), "context deadline exceeded") {
						isTimeout = true
						speed := result.GetAvgSpeed()
						if speed > 0 {
							wt.sendLog("info", fmt.Sprintf("Worker %d: %s 超时，使用部分结�?(%.2f MB/s)", workerID+1, proxy.Name, speed))
						} else {
							wt.sendLog("info", fmt.Sprintf("Worker %d: %s 超时，无可用数据", workerID+1, proxy.Name))
						}
						
						// Send result with whatever data we have
						wt.sendResult(proxy.Name, result, true)
						results <- result
						continue
					}
					
					wt.sendLog("error", fmt.Sprintf("Worker %d failed to test %s: %v", workerID+1, proxy.Name, err))
					
					// Send failed result
					wt.sendResult(proxy.Name, result, false)
					errors <- err
					continue
				}
				
				wt.results.Store(proxy.Name, result)
				
				// Initial ping
				if wt.EnablePing {
					delay, err := wt.clash.TestProxyDelay(proxy.Name, uint16(wt.PingTimeout), "")
					if err == nil {
						result.AddPing(delay)
					}
				}
				
				wt.sendResult(proxy.Name, result, isTimeout)
				results <- result
			}
		}(w)
	}
	
	// Queue all proxies
	for i, proxy := range wt.proxies {
		proxyQueue <- proxyJob{proxy: proxy, index: i}
	}
	close(proxyQueue)
	
	// Wait for completion or stop signal
	completedCount := 0
	errorCount := 0
	for completedCount+errorCount < len(wt.proxies) {
		select {
		case <-results:
			completedCount++
		case <-errors:
			errorCount++
		case <-wt.StopChan:
			wt.sendLog("info", fmt.Sprintf("Received stop signal, stopping all workers (completed: %d, failed: %d)", completedCount, errorCount))
			// Drain remaining results to avoid goroutine leak
			go func() {
				for completedCount+errorCount < len(wt.proxies) {
					select {
					case <-results:
						completedCount++
					case <-errors:
						errorCount++
					case <-time.After(100 * time.Millisecond):
						return
					}
				}
			}()
			return nil
		}
	}
	
	wt.sendLog("info", fmt.Sprintf("Completed: %d, Failed: %d", completedCount, errorCount))
	return nil
}

func (wt *WebTask) pingMonitorWithBroadcast() {
	ticker := time.NewTicker(time.Duration(wt.PingInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-wt.pingStop:
			return
		case <-ticker.C:
			wt.results.Range(func(name string, result *Result) bool {
				delay, err := wt.clash.TestProxyDelay(name, uint16(wt.PingTimeout), "")
				if err == nil {
					result.AddPing(delay)
					wt.sendPingUpdate(name, delay)
				}
				return true
			})
		}
	}
}

func (wt *WebTask) sendLog(level, message string) {
	if wt.broadcast != nil {
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "log",
			"data": map[string]interface{}{
				"level":   level,
				"message": message,
				"time":    time.Now().Format("15:04:05"),
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

func (wt *WebTask) sendProgress(current, total int, proxyName, status string) {
	if wt.broadcast != nil {
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "progress",
			"data": map[string]interface{}{
				"current":    current,
				"total":      total,
				"proxy_name": proxyName,
				"status":     status,
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

func (wt *WebTask) sendResult(proxyName string, result *Result, isTimeout bool) {
	if wt.broadcast != nil {
		// Calculate speed min/max from downloads
		var speedMin, speedMax float64
		for _, dl := range result.downloads {
			if dl != nil && dl.RespTime > 0 {
				speed := float64(dl.TotalSize) / dl.RespTime.Seconds() / 1024 / 1024
				if speedMin == 0 || speed < speedMin {
					speedMin = speed
				}
				if speed > speedMax {
					speedMax = speed
				}
			}
		}
		
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "result",
			"data": map[string]interface{}{
				"proxy_name": proxyName,
				"speed":      result.GetAvgSpeed(),
				"speed_min":  speedMin,
				"speed_max":  speedMax,
				"ping_min":   result.pingMin,
				"ping_max":   result.pingMax,
				"ping_avg":   result.GetPingAvg(),
				"ping_count": result.pingCount,
				"is_timeout": isTimeout,
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

func (wt *WebTask) sendPingUpdate(proxyName string, delay uint16) {
	if wt.broadcast != nil {
		// Use non-blocking send to prevent panic if channel is closed
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "ping_update",
			"data": map[string]interface{}{
				"proxy_name": proxyName,
				"delay":      delay,
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

func (wt *WebTask) sendSpeedUpdate(proxyName string, speed float64) {
	if wt.broadcast != nil {
		// Use non-blocking send to prevent panic if channel is closed
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "speed_update",
			"data": map[string]interface{}{
				"proxy_name": proxyName,
				"speed":      speed,
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

func (wt *WebTask) monitorSpeed(ctx context.Context, proxyName string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if result, ok := wt.results.Load(proxyName); ok {
				speed := result.GetAvgSpeed()
				if speed > 0 {
					wt.sendSpeedUpdate(proxyName, speed)
				}
			}
		}
	}
}

func (wt *WebTask) TestWithTimeout(proxy *clash.Proxy, timeout time.Duration, preResult *Result) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	resultChan := make(chan *Result, 1)
	errChan := make(chan error, 1)
	
	go func() {
		// Use the pre-created result if provided
		var result *Result
		var err error
		
		if preResult != nil {
			// Set global proxy
			err = wt.clash.SetGlobalProxy(proxy.Name)
			if err != nil {
				errChan <- err
				return
			}
			
			// Run download with pre-created result
			err = wt.Down(preResult)
			if err != nil {
				errChan <- err
				return
			}
			result = preResult
		} else {
			// Fallback to normal Test
			result, err = wt.Test(proxy)
			if err != nil {
				errChan <- err
				return
			}
		}
		
		resultChan <- result
	}()
	
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		// Return partial result even on error
		if preResult != nil {
			return preResult, err
		}
		return nil, err
	case <-ctx.Done():
		// Timeout occurred, return partial result
		if preResult != nil {
			return preResult, ctx.Err()
		}
		// Try to get from stored results
		if result, ok := wt.results.Load(proxy.Name); ok {
			return result, ctx.Err()
		}
		return nil, ctx.Err()
	}
}

func (wt *WebTask) sendProxies() {
	if wt.broadcast != nil {
		proxies := make([]map[string]interface{}, len(wt.proxies))
		for i, p := range wt.proxies {
			proxies[i] = map[string]interface{}{
				"name": p.Name,
				"type": p.Type,
			}
		}
		
		// 获取订阅信息
		var providerNames []string
		if wt.providers != nil {
			for name := range wt.providers.Providers {
				providerNames = append(providerNames, name)
			}
			slog.Info("sending %d providers to web UI", len(providerNames))
		} else {
			slog.Info("no providers to send to web UI")
		}
		
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "proxies",
			"data": map[string]interface{}{
				"proxies":   proxies,
				"total":     len(proxies),
				"providers": providerNames,
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

func (wt *WebTask) sendResults() {
	if wt.broadcast != nil {
		results := make([]map[string]interface{}, 0)
		for _, proxy := range wt.proxies {
			if result, ok := wt.results.Load(proxy.Name); ok {
				results = append(results, map[string]interface{}{
					"proxy_name": proxy.Name,
					"proxy_type": proxy.Type,
					"speed":      result.GetAvgSpeed(),
					"ping_min":   result.pingMin,
					"ping_max":   result.pingMax,
					"ping_avg":   result.GetPingAvg(),
					"ping_count": result.pingCount,
				})
			}
		}
		select {
		case wt.broadcast <- map[string]interface{}{
			"type": "all_results",
			"data": map[string]interface{}{
				"results": results,
			},
		}:
		default:
			// Channel full or closed, skip this update
		}
	}
}

// Helper functions
func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getInt(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return defaultVal
}

func getBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return []string{}
}


// Override progress method to send realtime speed updates
func (wt *WebTask) progress(bar *util.ProgressBar, proxyName string) common.DownloadBodyFunc {
	return func(body io.ReadCloser, size int64) error {
		if bar != nil {
			defer bar.Finish()
			bar.SetTotal(size)
			_, err := io.Copy(io.Discard, bar.NewProxyReader(body))
			return err
		}
		
		// Log mode: periodically output progress
		buf := make([]byte, 32*1024) // 32KB chunks
		var downloaded int64
		startTime := time.Now()
		lastLog := time.Now()
		
		for {
			n, err := body.Read(buf)
			if n > 0 {
				downloaded += int64(n)
				
				// Log every 2 seconds
				if time.Since(lastLog) >= 2*time.Second {
					elapsed := time.Since(startTime).Seconds()
					speed := float64(downloaded) / elapsed / 1024 / 1024
					percent := float64(downloaded) * 100 / float64(size)
					
					// Send realtime speed update
					wt.sendSpeedUpdate(proxyName, speed)
					
					slog.Info("downloading: %.1f%% (%.2f MB / %.2f MB) @ %.2f MB/s",
						percent,
						float64(downloaded)/1024/1024,
						float64(size)/1024/1024,
						speed)
					lastLog = time.Now()
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
}


// Override Down method to use custom progress with realtime updates
func (wt *WebTask) Down(result *Result) error {
	var down downFunc
	var err error
	
	switch wt.Download {
	case "cloudflare":
		down, err = wt.downCloudflare(result)
	case "speedtest":
		down, err = wt.downSpeedtest(result)
	case "fast":
		down, err = wt.downFast(result)
	default:
		return fmt.Errorf("unknown download type: %s", wt.Download)
	}
	if err != nil {
		return err
	}

	// Use log mode instead of progress bars for concurrent testing
	wg := sync.WaitGroup{}
	wg.Add(wt.Threads)

	for i := 0; i < wt.Threads; i++ {
		go func(i int) { down(i, nil, &wg) }(i)
	}

	wg.Wait()

	return nil
}

func (wt *WebTask) downCloudflare(result *Result) (downFunc, error) {
	cli := cloudflare.NewClient().WithProxy(wt.ClashProxy)

	cfg, err := cli.GetConfig()
	if err != nil {
		return nil, err
	}
	result.Ip = cfg.Ip
	result.Country = cfg.Country
	result.Lat = cfg.Lat
	result.Lon = cfg.Lon

	fn := func(i int, bar *util.ProgressBar, wg *sync.WaitGroup) {
		defer wg.Done()
		res, err2 := cli.Download(wt.Size, wt.progress(bar, result.Proxy.Name))
		if err2 != nil {
			slog.Error("cloudflare error: %v", err2)
			return
		}
		result.SetDownload(i, res)
	}

	return fn, nil
}

func (wt *WebTask) downSpeedtest(result *Result) (downFunc, error) {
	cli := speedtest.NewClient().WithProxy(wt.ClashProxy)

	cfg, err := cli.GetConfig()
	if err != nil {
		return nil, err
	}
	result.Ip = cfg.Client.Ip
	result.Country = cfg.Client.Country
	result.Lat = cfg.Client.Lat
	result.Lon = cfg.Client.Lon

	servers, err := cli.GetServers()
	if err != nil {
		return nil, err
	}

	fn := func(i int, bar *util.ProgressBar, wg *sync.WaitGroup) {
		defer wg.Done()
		server := servers[i%len(servers)]
		res, err2 := cli.Download(server, wt.Size, wt.progress(bar, result.Proxy.Name))
		if err2 != nil {
			slog.Error("speedtest error: %v", err2)
			return
		}
		result.SetDownload(i, res)
	}

	return fn, nil
}

func (wt *WebTask) downFast(result *Result) (downFunc, error) {
	cli := fast.NewClient().WithProxy(wt.ClashProxy)

	cfg, err := cli.GetConfig()
	if err != nil {
		return nil, err
	}
	result.Ip = cfg.Client.Ip
	result.Country = cfg.Client.Location.Country
	result.City = cfg.Client.Location.City

	fn := func(i int, bar *util.ProgressBar, wg *sync.WaitGroup) {
		defer wg.Done()
		target := cfg.Targets[i%len(cfg.Targets)]
		res, err2 := cli.Download(target, wt.Size, wt.progress(bar, result.Proxy.Name))
		if err2 != nil {
			slog.Error("fast error: %v", err2)
			return
		}
		result.SetDownload(i, res)
	}

	return fn, nil
}
