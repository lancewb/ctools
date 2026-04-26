package network

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendHttpRequestHandlesDefaultsHeadersAndErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET method, got %s", r.Method)
		}
		if got := r.Header.Get("X-Test"); got != "ok" {
			t.Fatalf("expected header X-Test=ok, got %q", got)
		}
		w.Header().Set("X-Reply", "yes")
		_, _ = w.Write([]byte("pong"))
	}))
	defer server.Close()

	result := NewNetworkService().SendHttpRequest(RequestOption{
		URL:     server.URL,
		Headers: map[string]string{"X-Test": "ok"},
	})
	if result.Error != "" {
		t.Fatalf("SendHttpRequest failed: %s", result.Error)
	}
	if result.StatusCode != http.StatusOK || result.Body != "pong" {
		t.Fatalf("unexpected response: %+v", result)
	}
	if result.Headers["X-Reply"] != "yes" {
		t.Fatalf("expected response header to be preserved")
	}

	if got := NewNetworkService().SendHttpRequest(RequestOption{}); got.Error == "" {
		t.Fatalf("expected missing URL error")
	}
	if got := NewNetworkService().SendHttpRequest(RequestOption{URL: "example.com", Protocol: "ws"}); got.Error == "" {
		t.Fatalf("expected unsupported protocol error")
	}
}

func TestRequestCollectionsPersistence(t *testing.T) {
	t.Setenv("CTOOLS_CONFIG_DIR", t.TempDir())
	service := NewNetworkService()

	created := service.SaveReqCollection(CollectionItem{
		Name: " first ",
		Request: RequestOption{
			Method: "POST",
			URL:    "https://example.test",
		},
	})
	if len(created) != 1 || created[0].ID == "" {
		t.Fatalf("expected generated collection id, got %+v", created)
	}
	if created[0].Name != "first" {
		t.Fatalf("expected collection name to be trimmed")
	}

	created[0].Name = "updated"
	updated := service.SaveReqCollection(created[0])
	if len(updated) != 1 || updated[0].Name != "updated" {
		t.Fatalf("expected existing collection to be updated, got %+v", updated)
	}

	deleted := service.DeleteReqCollection(updated[0].ID)
	if len(deleted) != 0 {
		t.Fatalf("expected collection to be deleted, got %+v", deleted)
	}
}

func TestPingHistoryPersistence(t *testing.T) {
	t.Setenv("CTOOLS_CONFIG_DIR", t.TempDir())
	service := NewNetworkService()

	for i := 0; i < 55; i++ {
		service.AddPingHistory(fmt.Sprintf("192.168.%d", i))
	}
	service.AddPingHistory("192.168.10")

	history := service.GetPingHistory()
	if len(history) != 50 {
		t.Fatalf("expected 50 history entries, got %d", len(history))
	}
	if history[0] != "192.168.10" {
		t.Fatalf("expected duplicate entry to move to front, got %q", history[0])
	}

	history = service.RemovePingHistory("192.168.10")
	if len(history) != 49 || (len(history) > 0 && history[0] == "192.168.10") {
		t.Fatalf("expected history entry to be removed, got %+v", history)
	}

	service.ClearPingHistory()
	if got := service.GetPingHistory(); len(got) != 0 {
		t.Fatalf("expected history to be cleared, got %+v", got)
	}
}

func TestServerPersistenceAndValidation(t *testing.T) {
	t.Setenv("CTOOLS_CONFIG_DIR", t.TempDir())
	service := NewNetworkService()

	list := service.SaveServer(ServerConfig{
		Name: " box ",
		Host: " 127.0.0.1 ",
		User: " root ",
	})
	if len(list) != 1 || list[0].Port != "22" || list[0].AuthType != "password" {
		t.Fatalf("expected normalized server defaults, got %+v", list)
	}
	if list[0].Name != "box" || list[0].Host != "127.0.0.1" || list[0].User != "root" {
		t.Fatalf("expected server fields to be trimmed, got %+v", list[0])
	}

	inserted := service.SaveServer(ServerConfig{ID: "external", Name: "new", Host: "host", User: "user"})
	if len(inserted) != 2 {
		t.Fatalf("expected non-existing explicit id to be appended, got %+v", inserted)
	}

	status := service.CheckServerStatus(ServerConfig{})
	if status.IsOnline || status.Error == "" {
		t.Fatalf("expected validation error for missing host/user, got %+v", status)
	}

	deleted := service.DeleteServer(list[0].ID)
	if len(deleted) != 1 || deleted[0].ID != "external" {
		t.Fatalf("expected server to be deleted, got %+v", deleted)
	}
}

func TestNetworkToolsDNSPortsTCPAndPrometheus(t *testing.T) {
	service := NewNetworkService()

	dns := service.LookupDNS(DNSLookupRequest{Host: "localhost"})
	if dns.Host != "localhost" || len(dns.Addresses) == 0 {
		t.Fatalf("expected localhost DNS addresses, got %+v", dns)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 128)
				n, _ := c.Read(buf)
				if n > 0 {
					_, _ = c.Write([]byte("echo:" + string(buf[:n])))
				}
			}(conn)
		}
	}()
	port := listener.Addr().(*net.TCPAddr).Port
	scan := service.ScanPorts(PortScanRequest{Host: "127.0.0.1", Ports: fmt.Sprintf("%d,%d-%d", port, port+1, port+1), TimeoutMillis: 200})
	if len(scan) != 2 || !scan[0].Open || scan[0].Port != port {
		t.Fatalf("unexpected scan results: %+v", scan)
	}

	tcp := service.SendTCP(TCPClientRequest{Host: "127.0.0.1", Port: port, Payload: "ping", TimeoutMillis: 500})
	if !tcp.Connected || tcp.Response != "echo:ping" {
		t.Fatalf("unexpected tcp result: %+v", tcp)
	}
	decoded, err := base64.StdEncoding.DecodeString(tcp.ResponseBase64)
	if err != nil || string(decoded) != "echo:ping" {
		t.Fatalf("unexpected tcp response base64: %q err=%v", tcp.ResponseBase64, err)
	}

	promServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			t.Fatalf("unexpected prometheus path: %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, "# HELP ctools_test counter\nctools_test_total{job=\"unit\"} 42\n")
	}))
	defer promServer.Close()
	_, portText, err := net.SplitHostPort(strings.TrimPrefix(promServer.URL, "http://"))
	if err != nil {
		t.Fatalf("split prometheus url: %v", err)
	}
	var promPort int
	_, _ = fmt.Sscanf(portText, "%d", &promPort)
	prom := service.ScrapePrometheus(PrometheusScrapeRequest{Host: "127.0.0.1", Port: promPort, Path: "/metrics"})
	if prom.Error != "" || len(prom.Metrics) != 1 || prom.Metrics[0].Name != "ctools_test_total" || prom.Metrics[0].Value != 42 {
		t.Fatalf("unexpected prometheus result: %+v", prom)
	}
}
