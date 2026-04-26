package other

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"ctools/backend/crypto"
)

func TestSocks5ServerConnectsToTCPDestination(t *testing.T) {
	target, targetAddr := startEchoServer(t)
	defer target.Close()

	proxyPort := freeTCPPort(t)
	service := NewOtherService(crypto.NewCryptoService())
	status, err := service.StartSocks5Proxy(Socks5Config{ListenIP: "127.0.0.1", Port: proxyPort})
	if err != nil {
		t.Fatalf("StartSocks5Proxy failed: %v", err)
	}
	defer service.StopSocks5Proxy()
	if !status.Running {
		t.Fatalf("expected SOCKS server to be running")
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", proxyPort)), 2*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatalf("write greeting: %v", err)
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		t.Fatalf("read greeting reply: %v", err)
	}
	if reply[0] != 0x05 || reply[1] != 0x00 {
		t.Fatalf("unexpected greeting reply: %x", reply)
	}

	host, portText, err := net.SplitHostPort(targetAddr)
	if err != nil {
		t.Fatalf("split target addr: %v", err)
	}
	targetPort := uint16(mustAtoi(t, portText))
	request := []byte{0x05, 0x01, 0x00, 0x01}
	request = append(request, net.ParseIP(host).To4()...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, targetPort)
	request = append(request, portBytes...)
	if _, err := conn.Write(request); err != nil {
		t.Fatalf("write connect request: %v", err)
	}
	connectReply := make([]byte, 10)
	if _, err := io.ReadFull(conn, connectReply); err != nil {
		t.Fatalf("read connect reply: %v", err)
	}
	if connectReply[1] != 0x00 {
		t.Fatalf("expected SOCKS connect success, got %x", connectReply)
	}

	if _, err := conn.Write([]byte("hello")); err != nil {
		t.Fatalf("write proxied payload: %v", err)
	}
	echo := make([]byte, 5)
	if _, err := io.ReadFull(conn, echo); err != nil {
		t.Fatalf("read proxied payload: %v", err)
	}
	if string(echo) != "hello" {
		t.Fatalf("unexpected proxied response: %q", echo)
	}

	stopped, err := service.StopSocks5Proxy()
	if err != nil {
		t.Fatalf("StopSocks5Proxy failed: %v", err)
	}
	if stopped.Running {
		t.Fatalf("expected SOCKS server to stop")
	}
}

func TestGMSSLValidationAndStatus(t *testing.T) {
	service := NewOtherService(crypto.NewCryptoService())
	if _, err := service.StartGMSSLServer(GMSSLServerConfig{}); err == nil {
		t.Fatalf("expected StartGMSSLServer validation error")
	}
	if _, err := service.RunGMSSLClientTest(GMSSLClientConfig{}); err == nil {
		t.Fatalf("expected RunGMSSLClientTest validation error")
	}
	if status := service.GMSSLServerStatus(); status.Running {
		t.Fatalf("expected GMSSL server to be stopped")
	}
}

func startEchoServer(t *testing.T) (net.Listener, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen echo: %v", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}(conn)
		}
	}()
	return listener, listener.Addr().String()
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen free port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func mustAtoi(t *testing.T, text string) int {
	t.Helper()
	var value int
	if _, err := fmt.Sscanf(text, "%d", &value); err != nil {
		t.Fatalf("parse int %q: %v", text, err)
	}
	return value
}
