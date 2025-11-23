package network

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/google/uuid"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// RequestOption 前端传来的请求参数
type RequestOption struct {
	ID         string            `json:"id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`       // JSON string
	Protocol   string            `json:"protocol"`   // http, https
	TlsVersion string            `json:"tlsVersion"` // "", "1.1", "1.2", "1.3", "tlcp"
	Timeout    int               `json:"timeout"`    // 秒
}

// ResponseResult 返回给前端的结果
type ResponseResult struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	TimeCost   int64             `json:"timeCost"` // 毫秒
	Error      string            `json:"error"`
}

// CollectionItem 收藏夹项
type CollectionItem struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Request RequestOption `json:"request"`
}

// SendHttpRequest 发送 HTTP 请求
func (n *NetworkService) SendHttpRequest(opt RequestOption) ResponseResult {
	start := time.Now()

	// 1. 构造 Body (保持不变)
	var bodyReader io.Reader
	if opt.Body != "" {
		bodyReader = bytes.NewBufferString(opt.Body)
	}

	// 2. 创建 Request (保持不变)
	req, err := http.NewRequest(opt.Method, opt.URL, bodyReader)
	if err != nil {
		return ResponseResult{Error: "创建请求失败: " + err.Error()}
	}

	// 3. 设置 Headers (保持不变)
	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	// 4. 配置 Transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// 处理 TLS/TLCP
	if opt.TlsVersion == "tlcp" {
		// --- 修改点：使用 gotlcp 实现国密 TLCP ---
		transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 1. 建立基础 TCP 连接
			conn, err := net.DialTimeout(network, addr, 10*time.Second)
			if err != nil {
				return nil, err
			}

			// 2. 配置 TLCP
			// 注意：gotlcp 的 Config 结构通常位于 conf 包中
			tlcpConfig := &tlcp.Config{
				InsecureSkipVerify: true, // 跳过证书校验（模拟 Postman 行为）
			}

			// 3. 进行 TLCP 握手
			// gotlcp.Client 返回的是 *tlcp.Conn
			tlsConn := tlcp.Client(conn, tlcpConfig)

			// 必须手动 Handshake 以便尽早捕获错误，否则会在 Write 时才触发
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				conn.Close()
				return nil, fmt.Errorf("TLCP 握手失败: %v", err)
			}

			return tlsConn, nil
		}
	} else if opt.URL[0:5] == "https" {
		// 标准 TLS 配置 (保持不变)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		switch opt.TlsVersion {
		case "1.1":
			tlsConfig.MinVersion = tls.VersionTLS11
			tlsConfig.MaxVersion = tls.VersionTLS11
		case "1.2":
			tlsConfig.MinVersion = tls.VersionTLS12
			tlsConfig.MaxVersion = tls.VersionTLS12
		case "1.3":
			tlsConfig.MinVersion = tls.VersionTLS13
			tlsConfig.MaxVersion = tls.VersionTLS13
		}
		transport.TLSClientConfig = tlsConfig
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(opt.Timeout) * time.Second,
	}

	// 5. 发送请求 (保持不变)
	resp, err := client.Do(req)
	cost := time.Since(start).Milliseconds()

	if err != nil {
		return ResponseResult{Error: "请求发送失败: " + err.Error(), TimeCost: cost}
	}
	defer resp.Body.Close()

	// 6. 读取响应 (保持不变)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseResult{Error: "读取响应失败: " + err.Error(), TimeCost: cost}
	}

	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	return ResponseResult{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(respBody),
		TimeCost:   cost,
	}
}

// --- 收藏夹管理 (XDG) ---

const reqCollectionFile = "request_collections.json"

func (n *NetworkService) GetReqCollections() []CollectionItem {
	path := filepath.Join(filepath.Dir(n.getConfigPath()), reqCollectionFile) // 复用之前的路径逻辑
	data, err := os.ReadFile(path)
	if err != nil {
		return []CollectionItem{}
	}
	var list []CollectionItem
	json.Unmarshal(data, &list)
	return list
}

func (n *NetworkService) SaveReqCollection(item CollectionItem) []CollectionItem {
	list := n.GetReqCollections()

	// 如果 ID 为空，生成新 ID
	if item.ID == "" {
		item.ID = uuid.New().String()
		list = append(list, item)
	} else {
		// 更新现有
		found := false
		for i, v := range list {
			if v.ID == item.ID {
				list[i] = item
				found = true
				break
			}
		}
		if !found {
			list = append(list, item)
		}
	}

	n.saveReqList(list)
	return list
}

func (n *NetworkService) DeleteReqCollection(id string) []CollectionItem {
	list := n.GetReqCollections()
	newList := []CollectionItem{}
	for _, v := range list {
		if v.ID != id {
			newList = append(newList, v)
		}
	}
	n.saveReqList(newList)
	return newList
}

func (n *NetworkService) saveReqList(list []CollectionItem) {
	path := filepath.Join(filepath.Dir(n.getConfigPath()), reqCollectionFile)
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(path, data, 0644)
}
