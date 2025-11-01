package job

import (
	"sync"
	"time"

	"github.com/starudream/clash-speedtest/api/clash"
	"github.com/starudream/clash-speedtest/api/common"
)

type Result struct {
	Proxy *clash.Proxy

	Ip      string
	Country string
	City    string
	Lat     string
	Lon     string

	threads   int
	total     *common.DownloadResult
	downloads []*common.DownloadResult
	
	// Ping statistics
	pingCount int
	pingMin   uint16 // milliseconds
	pingMax   uint16 // milliseconds
	pingSum   uint64 // milliseconds
	
	mu        sync.Mutex
}

func (t *Result) SetDownload(i int, v *common.DownloadResult) {
	if v == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.downloads[i] = v
	t.total.TotalSize += v.TotalSize
	t.total.ConnTime += v.ConnTime
	t.total.RespTime += v.RespTime
}

// AddPing adds a ping result to statistics
func (t *Result) AddPing(delay uint16) {
	if delay == 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.pingCount++
	t.pingSum += uint64(delay)
	
	if t.pingMin == 0 || delay < t.pingMin {
		t.pingMin = delay
	}
	if delay > t.pingMax {
		t.pingMax = delay
	}
}

// GetPingAvg returns average ping in milliseconds
func (t *Result) GetPingAvg() uint16 {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.pingCount == 0 {
		return 0
	}
	return uint16(t.pingSum / uint64(t.pingCount))
}

// GetAvgSpeed returns average download speed in MB/s
// 计算所有并发线程的平均速度：总下载量 / 最长耗时
func (t *Result) GetAvgSpeed() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// 找出最长的下载时间（因为是并发的）
	var maxTime time.Duration
	for _, dl := range t.downloads {
		if dl != nil && dl.RespTime > maxTime {
			maxTime = dl.RespTime
		}
	}
	
	if maxTime.Seconds() == 0 {
		return 0
	}
	
	// 平均速度 = 总下载量 / 最长耗时
	return float64(t.total.TotalSize) / maxTime.Seconds() / 1024 / 1024
}


func (t *Task) Test(proxy *clash.Proxy) (*Result, error) {
	err := t.clash.SetGlobalProxy(proxy.Name)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Proxy:     proxy,
		threads:   t.Threads,
		total:     &common.DownloadResult{},
		downloads: make([]*common.DownloadResult, t.Threads),
	}

	err = t.Down(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
