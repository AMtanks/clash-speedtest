package job

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/starudream/go-lib/core/v2/slog"
	"github.com/starudream/go-lib/core/v2/utils/maputil"

	"github.com/starudream/clash-speedtest/api/clash"
)

type WebTask struct {
	*Task
	broadcast chan<- interface{}
	speedUpdates map[string]*SpeedTracker
	speedMutex sync.RWMutex
}

type SpeedTracker struct {
	lastSpeed float64
	lastUpdate time.Time
}

func NewTaskFromConfig(config map[string]interface{}, broadcast chan<- interface{}) *WebTask {
	task := &Task{
		ClashAddr:       getString(config, "clash_addr", "http://127.0.0.1:9090"),
		ClashSecret:     getString(config, "clash_secret", ""),
		ClashProxy:      getString(config, "clash_proxy", ""),
		Size:            getInt(config, "size", 10),
		Threads:         getInt(config, "threads", 1),
		Download:        getString(config, "download", "cloudflare"),
		Includes:        getStringSlice(config, "includes"),
		Excludes:        getStringSlice(config, "excludes"),
		Confirm:         true,
		Output:          getString(config, "output", "output"),
		Timeout:         getInt(config, "timeout", 60),
		EnablePing:      getBool(config, "ping", false),
		PingInterval:    getInt(config, "ping_interval", 60),
		PingTimeout:     getInt(config, "ping_timeout", 5000),
		OutputFormat:    getString(config, "format", "txt"),
		ConcurrentNodes: getInt(config, "concurrent", 1),
		results:         maputil.SyncMap[string, *Result]{},
	}
	
	return &WebTask{
		Task:         task,
		broadcast:    broadcast,
		speedUpdates: make(map[string]*SpeedTracker),
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
	
	// Start ping monitoring if enabled
	if wt.EnablePing {
		wt.pingStop = make(chan struct{})
		go wt.pingMonitorWithBroadcast()
		wt.sendLog("info", fmt.Sprintf("Ping monitoring enabled (interval: %ds)", wt.PingInterval))
	}
	
	defer func() {
		if wt.EnablePing && wt.pingStop != nil {
			close(wt.pingStop)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	
	// Run speedtest with progress updates
	wt.sendLog("info", fmt.Sprintf("Starting speedtest with %d concurrent workers...", wt.ConcurrentNodes))
	err = wt.runSpeedtestWithProgress()
	if err != nil {
		wt.sendLog("error", fmt.Sprintf("Speedtest failed: %v", err))
		return err
	}
	
	wt.sendLog("info", "Speedtest completed, sending results...")
	wt.sendResults()
	
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
				proxy := job.proxy
				
				wt.sendProgress(job.index+1, len(wt.proxies), proxy.Name, "testing")
				wt.sendLog("info", fmt.Sprintf("Worker %d testing: %s", workerID+1, proxy.Name))
				
				// Start speed monitoring for this proxy
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(wt.Timeout)*time.Second)
				defer cancel()
				
				go wt.monitorSpeed(ctx, proxy.Name)
				
				result, err := wt.TestWithTimeout(proxy, time.Duration(wt.Timeout)*time.Second)
				cancel() // Stop speed monitoring
				
				if err != nil {
					// Check if it's a timeout error
					if ctx.Err() == context.DeadlineExceeded || strings.Contains(err.Error(), "context deadline exceeded") {
						wt.sendLog("error", fmt.Sprintf("Worker %d timeout for %s, using average speed", workerID+1, proxy.Name))
						// Use average speed from partial results
						if result != nil && result.GetAvgSpeed() > 0 {
							wt.results.Store(proxy.Name, result)
							wt.sendResult(proxy.Name, result)
							results <- result
							continue
						}
					}
					wt.sendLog("error", fmt.Sprintf("Worker %d failed to test %s: %v", workerID+1, proxy.Name, err))
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
				
				wt.sendResult(proxy.Name, result)
				results <- result
			}
		}(w)
	}
	
	// Queue all proxies
	for i, proxy := range wt.proxies {
		proxyQueue <- proxyJob{proxy: proxy, index: i}
	}
	close(proxyQueue)
	
	// Wait for completion
	completedCount := 0
	errorCount := 0
	for completedCount+errorCount < len(wt.proxies) {
		select {
		case <-results:
			completedCount++
		case <-errors:
			errorCount++
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
		wt.broadcast <- map[string]interface{}{
			"type": "log",
			"data": map[string]interface{}{
				"level":   level,
				"message": message,
				"time":    time.Now().Format("15:04:05"),
			},
		}
	}
}

func (wt *WebTask) sendProgress(current, total int, proxyName, status string) {
	if wt.broadcast != nil {
		wt.broadcast <- map[string]interface{}{
			"type": "progress",
			"data": map[string]interface{}{
				"current":    current,
				"total":      total,
				"proxy_name": proxyName,
				"status":     status,
			},
		}
	}
}

func (wt *WebTask) sendResult(proxyName string, result *Result) {
	if wt.broadcast != nil {
		wt.broadcast <- map[string]interface{}{
			"type": "result",
			"data": map[string]interface{}{
				"proxy_name": proxyName,
				"speed":      result.GetAvgSpeed(),
				"ping_min":   result.pingMin,
				"ping_max":   result.pingMax,
				"ping_avg":   result.GetPingAvg(),
				"ping_count": result.pingCount,
			},
		}
	}
}

func (wt *WebTask) sendPingUpdate(proxyName string, delay uint16) {
	if wt.broadcast != nil {
		wt.broadcast <- map[string]interface{}{
			"type": "ping_update",
			"data": map[string]interface{}{
				"proxy_name": proxyName,
				"delay":      delay,
			},
		}
	}
}

func (wt *WebTask) sendSpeedUpdate(proxyName string, speed float64) {
	if wt.broadcast != nil {
		wt.broadcast <- map[string]interface{}{
			"type": "speed_update",
			"data": map[string]interface{}{
				"proxy_name": proxyName,
				"speed":      speed,
			},
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

func (wt *WebTask) TestWithTimeout(proxy *clash.Proxy, timeout time.Duration) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	resultChan := make(chan *Result, 1)
	errChan := make(chan error, 1)
	
	go func() {
		result, err := wt.Test(proxy)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()
	
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		// Timeout occurred, try to get partial result
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
		wt.broadcast <- map[string]interface{}{
			"type": "proxies",
			"data": map[string]interface{}{
				"proxies": proxies,
				"total":   len(proxies),
			},
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
		wt.broadcast <- map[string]interface{}{
			"type": "all_results",
			"data": map[string]interface{}{
				"results": results,
			},
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
