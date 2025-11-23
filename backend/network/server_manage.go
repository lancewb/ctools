package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// --- 数据结构 ---

type ServerConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	AuthType string `json:"authType"` // "password" or "key"
	Password string `json:"password"`
	KeyPath  string `json:"keyPath"`
}

type ServerStatus struct {
	ID          string   `json:"id"`
	IsOnline    bool     `json:"isOnline"`
	Error       string   `json:"error"`
	CPUModel    string   `json:"cpuModel"`
	CPUUsage    string   `json:"cpuUsage"` // e.g., "15%"
	RAMTotal    string   `json:"ramTotal"` // e.g., "16G"
	RAMUsed     string   `json:"ramUsed"`  // e.g., "4G"
	RAMPercent  float64  `json:"ramPercent"`
	DiskSize    string   `json:"diskSize"`
	DiskUsed    string   `json:"diskUsed"`
	DiskPercent float64  `json:"diskPercent"`
	PCIDevices  []string `json:"pciDevices"` // 网卡和加密卡列表
}

// --- 持久化管理 ---

const serverListFile = "server_list.json"

func (n *NetworkService) GetServerList() []ServerConfig {
	path := filepath.Join(filepath.Dir(n.getConfigPath()), serverListFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return []ServerConfig{}
	}
	var list []ServerConfig
	json.Unmarshal(data, &list)
	return list
}

func (n *NetworkService) SaveServer(server ServerConfig) []ServerConfig {
	list := n.GetServerList()

	// 生成ID
	if server.ID == "" {
		server.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		list = append(list, server)
	} else {
		// 更新
		for i, v := range list {
			if v.ID == server.ID {
				list[i] = server
				break
			}
		}
	}
	n.saveServerFile(list)
	return list
}

func (n *NetworkService) DeleteServer(id string) []ServerConfig {
	list := n.GetServerList()
	newList := []ServerConfig{}
	for _, v := range list {
		if v.ID != id {
			newList = append(newList, v)
		}
	}
	n.saveServerFile(newList)
	return newList
}

func (n *NetworkService) saveServerFile(list []ServerConfig) {
	path := filepath.Join(filepath.Dir(n.getConfigPath()), serverListFile)
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(path, data, 0644)
}

// --- SSH 核心逻辑 ---

// CheckServerStatus 连接服务器并获取详细信息
func (n *NetworkService) CheckServerStatus(config ServerConfig) ServerStatus {
	status := ServerStatus{ID: config.ID, IsOnline: false}

	// 1. 配置 SSH Client
	authMethods := []ssh.AuthMethod{}
	if config.AuthType == "key" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			status.Error = "读取私钥失败"
			return status
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			status.Error = "解析私钥失败"
			return status
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略 Host Key 检查 (简化模式)
		Timeout:         5 * time.Second,
	}

	// 2. 连接
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		status.Error = fmt.Sprintf("连接失败: %v", err)
		return status
	}
	defer client.Close()

	status.IsOnline = true

	// 3. 执行命令获取信息
	// 我们尽量合并命令以减少网络交互，或者分条执行

	// A. CPU 型号
	status.CPUModel = runCmd(client, "grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs")

	// B. CPU 使用率 (使用 top -bn1 简单获取 idle 值反推)
	// 注意：不同发行版 top 输出格式可能不同，这里使用比较通用的 awk 解析
	// 另一种方法是读 /proc/stat 两次，但这里为了速度取瞬时值
	idleStr := runCmd(client, "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/'")
	if idleStr != "" {
		// 简单处理：100 - idle = usage (需要在前端或这里转 float，这里偷懒直接返回字符串描述)
		status.CPUUsage = fmt.Sprintf("Load: %s", runCmd(client, "uptime | awk -F'load average:' '{ print $2 }' | xargs"))
	}

	// C. 内存 (Free -m)
	// 输出格式: Mem: 15000 4000 ...
	ramOut := runCmd(client, "free -m | grep Mem | awk '{print $2,$3}'")
	if parts := strings.Fields(ramOut); len(parts) >= 2 {
		total := parseFloats(parts[0])
		used := parseFloats(parts[1])
		status.RAMTotal = fmt.Sprintf("%.1f GB", total/1024)
		status.RAMUsed = fmt.Sprintf("%.1f GB", used/1024)
		if total > 0 {
			status.RAMPercent = (used / total) * 100
		}
	}

	// D. 硬盘 (df -h /)
	diskOut := runCmd(client, "df -h / | tail -1 | awk '{print $2,$3,$5}'")
	if parts := strings.Fields(diskOut); len(parts) >= 3 {
		status.DiskSize = parts[0]
		status.DiskUsed = parts[1]
		// 去掉 %
		usageStr := strings.TrimSuffix(parts[2], "%")
		status.DiskPercent = parseFloats(usageStr)
	}

	// E. PCI 设备 (过滤 网卡 和 加密卡)
	// 关键词: Ethernet, Network, Crypto, Accelerator
	pciCmd := "lspci | grep -Ei 'Ethernet|Network|Crypto|Accelerator'"
	pciOut := runCmd(client, pciCmd)
	if pciOut != "" {
		lines := strings.Split(pciOut, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				// 截取有用信息，去掉前面的 00:00.0 ID
				status.PCIDevices = append(status.PCIDevices, line)
			}
		}
	} else {
		// 如果没有 lspci，尝试 ip addr
		status.PCIDevices = append(status.PCIDevices, "无 lspci 命令，仅列出接口:")
		ipOut := runCmd(client, "ip -o link show | awk -F': ' '{print $2}'")
		status.PCIDevices = append(status.PCIDevices, strings.Split(ipOut, "\n")...)
	}

	return status
}

// 辅助：执行单条命令
func runCmd(client *ssh.Client, cmd string) string {
	session, err := client.NewSession()
	if err != nil {
		return ""
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		return ""
	}
	return strings.TrimSpace(b.String())
}

func parseFloats(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
