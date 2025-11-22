package network

import (
	"encoding/json"
	"fmt"
	"github.com/go-ping/ping"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

type PingResult struct {
	ID      int   `json:"id"`
	Latency int64 `json:"latency"`
}

// PingSubnet perform concurrent ping scans on a given subnet.
func (a *NetworkService) PingSubnet(subnet string) []PingResult {
	var results []PingResult
	var wg sync.WaitGroup
	resultChan := make(chan PingResult, 254)

	// On Windows, the Pinger must be executed with elevated privileges.
	if runtime.GOOS == "windows" {
		pinger, err := ping.NewPinger("127.0.0.1")
		if err != nil {
			fmt.Println("ERROR: ", err)
			return nil
		}
		pinger.SetPrivileged(true)
	}

	for i := 1; i <= 254; i++ {
		wg.Add(1)
		ip := fmt.Sprintf("%s.%d", subnet, i)
		go func(ip string, id int) {
			defer wg.Done()
			pinger, err := ping.NewPinger(ip)
			if err != nil {
				resultChan <- PingResult{ID: id, Latency: 5000}
				return
			}
			pinger.SetPrivileged(true)
			pinger.Count = 3
			pinger.Timeout = 5 * time.Second

			err = pinger.Run() // Blocks until finished.
			if err != nil {
				resultChan <- PingResult{ID: id, Latency: 5000}
				return
			}
			stats := pinger.Statistics()
			if stats.PacketsRecv > 0 {
				resultChan <- PingResult{ID: id, Latency: stats.AvgRtt.Milliseconds()}
			} else {
				resultChan <- PingResult{ID: id, Latency: 5000}
			}

		}(ip, i)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		results = append(results, result)
	}

	// Sort by ID to ensure consistent display on the front end
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
	return results
}

// --- 历史记录管理 (XDG规范) ---

const historyFileName = "ping_history.json"

// 获取配置路径
func (a *NetworkService) getConfigPath() string {
	configDir, err := os.UserConfigDir() // Windows: AppData/Roaming, Linux: .config
	if err != nil {
		return "."
	}
	appDir := filepath.Join(configDir, "WailsToolbox") // 你的应用名
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		os.MkdirAll(appDir, 0755)
	}
	return filepath.Join(appDir, historyFileName)
}

// GetPingHistory 获取历史记录
func (a *NetworkService) GetPingHistory() []string {
	path := a.getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}
	var history []string
	json.Unmarshal(data, &history)
	return history
}

// AddPingHistory 添加历史记录 (保留最近50条)
func (a *NetworkService) AddPingHistory(subnet string) []string {
	history := a.GetPingHistory()

	// 去重：如果已存在，先删除旧的
	newHistory := []string{}
	for _, h := range history {
		if h != subnet {
			newHistory = append(newHistory, h)
		}
	}

	// 插入到最前面
	newHistory = append([]string{subnet}, newHistory...)

	// 截取前50条
	if len(newHistory) > 50 {
		newHistory = newHistory[:50]
	}

	a.saveHistory(newHistory)
	return newHistory
}

// RemovePingHistory 删除单条
func (a *NetworkService) RemovePingHistory(subnet string) []string {
	history := a.GetPingHistory()
	newHistory := []string{}
	for _, h := range history {
		if h != subnet {
			newHistory = append(newHistory, h)
		}
	}
	a.saveHistory(newHistory)
	return newHistory
}

// ClearPingHistory 清空
func (a *NetworkService) ClearPingHistory() {
	a.saveHistory([]string{})
}

func (a *NetworkService) saveHistory(history []string) {
	data, _ := json.Marshal(history)
	os.WriteFile(a.getConfigPath(), data, 0644)
}
