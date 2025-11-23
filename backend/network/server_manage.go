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

// --- Data Structures ---

// ServerConfig represents the configuration for a remote server connection.
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

// ServerStatus contains real-time status information of a server.
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
	PCIDevices  []string `json:"pciDevices"` // Network and crypto cards
}

// --- Persistence Management ---

const serverListFile = "server_list.json"

// GetServerList retrieves the list of saved servers.
//
// Returns a slice of ServerConfig.
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

// SaveServer saves or updates a server configuration.
//
// server: The ServerConfig to save.
// Returns the updated list of servers.
func (n *NetworkService) SaveServer(server ServerConfig) []ServerConfig {
	list := n.GetServerList()

	// Generate ID
	if server.ID == "" {
		server.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		list = append(list, server)
	} else {
		// Update
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

// DeleteServer removes a server configuration by its ID.
//
// id: The ID of the server to remove.
// Returns the updated list of servers.
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

// --- SSH Core Logic ---

// CheckServerStatus connects to the server via SSH and retrieves detailed status information.
//
// config: The ServerConfig containing connection details.
// Returns a ServerStatus struct with system metrics.
func (n *NetworkService) CheckServerStatus(config ServerConfig) ServerStatus {
	status := ServerStatus{ID: config.ID, IsOnline: false}

	// 1. Configure SSH Client
	authMethods := []ssh.AuthMethod{}
	if config.AuthType == "key" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			status.Error = "Failed to read private key"
			return status
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			status.Error = "Failed to parse private key"
			return status
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Simplified: ignore host key check
		Timeout:         5 * time.Second,
	}

	// 2. Connect
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		status.Error = fmt.Sprintf("Connection failed: %v", err)
		return status
	}
	defer client.Close()

	status.IsOnline = true

	// 3. Execute commands to get info
	// Merge commands or run separately

	// A. CPU Model
	status.CPUModel = runCmd(client, "grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs")

	// B. CPU Usage
	idleStr := runCmd(client, "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/'")
	if idleStr != "" {
		status.CPUUsage = fmt.Sprintf("Load: %s", runCmd(client, "uptime | awk -F'load average:' '{ print $2 }' | xargs"))
	}

	// C. Memory (Free -m)
	// Output: Mem: 15000 4000 ...
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

	// D. Disk (df -h /)
	diskOut := runCmd(client, "df -h / | tail -1 | awk '{print $2,$3,$5}'")
	if parts := strings.Fields(diskOut); len(parts) >= 3 {
		status.DiskSize = parts[0]
		status.DiskUsed = parts[1]
		// Remove %
		usageStr := strings.TrimSuffix(parts[2], "%")
		status.DiskPercent = parseFloats(usageStr)
	}

	// E. PCI Devices (Filter Network and Crypto)
	pciCmd := "lspci | grep -Ei 'Ethernet|Network|Crypto|Accelerator'"
	pciOut := runCmd(client, pciCmd)
	if pciOut != "" {
		lines := strings.Split(pciOut, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				// Trim leading ID
				status.PCIDevices = append(status.PCIDevices, line)
			}
		}
	} else {
		// Fallback to ip addr if lspci is missing
		status.PCIDevices = append(status.PCIDevices, "No lspci, listing interfaces:")
		ipOut := runCmd(client, "ip -o link show | awk -F': ' '{print $2}'")
		status.PCIDevices = append(status.PCIDevices, strings.Split(ipOut, "\n")...)
	}

	return status
}

// runCmd executes a single command via SSH.
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
