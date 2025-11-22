package network

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type PingResult struct {
	ID      int   `json:"id"`
	Latency int64 `json:"latency"`
}

// PingSubnet 模拟群Ping，返回254个结果
// 真实场景这里应该并发去Ping，这里为了演示UI效果做随机模拟
func (a *NetworkService) PingSubnet(subnet string) []PingResult {
	// 模拟网络延迟，稍微睡一会，让前端有个Loading的感觉
	time.Sleep(500 * time.Millisecond)

	results := make([]PingResult, 254)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 254; i++ {
		// 模拟产生各种状态的延迟
		// 70% 概率 < 100ms (绿)
		// 15% 概率 100-1000ms (橘)
		// 10% 概率 1000-5000ms (红)
		// 5%  概率 > 5000ms (白/超时)

		chance := rand.Intn(100)
		var lat int64

		if chance < 70 {
			lat = int64(rand.Intn(99) + 1)
		} else if chance < 85 {
			lat = int64(rand.Intn(900) + 100)
		} else if chance < 95 {
			lat = int64(rand.Intn(4000) + 1000)
		} else {
			lat = 5000 // 超时
		}

		results[i] = PingResult{
			ID:      i + 1,
			Latency: lat,
		}
	}
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
