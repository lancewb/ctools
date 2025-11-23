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

// RequestOption defines the parameters for an HTTP/HTTPS request.
type RequestOption struct {
	ID         string            `json:"id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`       // JSON string
	Protocol   string            `json:"protocol"`   // http, https
	TlsVersion string            `json:"tlsVersion"` // "", "1.1", "1.2", "1.3", "tlcp"
	Timeout    int               `json:"timeout"`    // in seconds
}

// ResponseResult contains the response details of an HTTP request.
type ResponseResult struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	TimeCost   int64             `json:"timeCost"` // in milliseconds
	Error      string            `json:"error"`
}

// CollectionItem represents a saved HTTP request configuration.
type CollectionItem struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Request RequestOption `json:"request"`
}

// SendHttpRequest sends an HTTP or HTTPS request based on the provided options.
// It supports standard TLS and Chinese SM2/TLCP protocols.
//
// opt: The RequestOption struct containing URL, method, headers, etc.
// Returns a ResponseResult with the status, body, and timing information.
func (n *NetworkService) SendHttpRequest(opt RequestOption) ResponseResult {
	start := time.Now()

	// 1. Construct Body
	var bodyReader io.Reader
	if opt.Body != "" {
		bodyReader = bytes.NewBufferString(opt.Body)
	}

	// 2. Create Request
	req, err := http.NewRequest(opt.Method, opt.URL, bodyReader)
	if err != nil {
		return ResponseResult{Error: "Failed to create request: " + err.Error()}
	}

	// 3. Set Headers
	for k, v := range opt.Headers {
		req.Header.Set(k, v)
	}

	// 4. Configure Transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Handle TLS/TLCP
	if opt.TlsVersion == "tlcp" {
		// --- Use gotlcp for GM TLCP ---
		transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 1. Basic TCP Connection
			conn, err := net.DialTimeout(network, addr, 10*time.Second)
			if err != nil {
				return nil, err
			}

			// 2. Configure TLCP
			tlcpConfig := &tlcp.Config{
				InsecureSkipVerify: true, // Skip cert verification (Postman-like)
			}

			// 3. TLCP Handshake
			// gotlcp.Client returns *tlcp.Conn
			tlsConn := tlcp.Client(conn, tlcpConfig)

			// Handshake manually to catch errors early
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				conn.Close()
				return nil, fmt.Errorf("TLCP handshake failed: %v", err)
			}

			return tlsConn, nil
		}
	} else if opt.URL[0:5] == "https" {
		// Standard TLS Configuration
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

	// 5. Send Request
	resp, err := client.Do(req)
	cost := time.Since(start).Milliseconds()

	if err != nil {
		return ResponseResult{Error: "Request failed: " + err.Error(), TimeCost: cost}
	}
	defer resp.Body.Close()

	// 6. Read Response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseResult{Error: "Failed to read response: " + err.Error(), TimeCost: cost}
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

// --- Collection Management (XDG) ---

const reqCollectionFile = "request_collections.json"

// GetReqCollections retrieves the saved request collection.
//
// Returns a slice of CollectionItem.
func (n *NetworkService) GetReqCollections() []CollectionItem {
	path := filepath.Join(filepath.Dir(n.getConfigPath()), reqCollectionFile) // Reuse config path logic
	data, err := os.ReadFile(path)
	if err != nil {
		return []CollectionItem{}
	}
	var list []CollectionItem
	json.Unmarshal(data, &list)
	return list
}

// SaveReqCollection saves a request to the collection.
//
// item: The CollectionItem to save.
// Returns the updated collection list.
func (n *NetworkService) SaveReqCollection(item CollectionItem) []CollectionItem {
	list := n.GetReqCollections()

	// Generate new ID if empty
	if item.ID == "" {
		item.ID = uuid.New().String()
		list = append(list, item)
	} else {
		// Update existing
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

// DeleteReqCollection removes a request from the collection.
//
// id: The ID of the item to remove.
// Returns the updated collection list.
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
