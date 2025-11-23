package other

import (
	"context"
	"ctools/backend/crypto"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

type OtherService struct {
	ctx    context.Context
	crypto *crypto.CryptoService

	mu          sync.Mutex
	socksServer *socks5Server
	gmServer    *gmsslServer
}

func NewOtherService(cryptoSvc *crypto.CryptoService) *OtherService {
	return &OtherService{
		crypto: cryptoSvc,
	}
}

func (s *OtherService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SOCKS5

type Socks5Config struct {
	ListenIP string `json:"listenIp"`
	Port     int    `json:"port"`
}

type Socks5Status struct {
	Running            bool   `json:"running"`
	Address            string `json:"address"`
	ActiveConnections  int64  `json:"activeConnections"`
	Error              string `json:"error"`
	LastControlMessage string `json:"lastControlMessage"`
}

type socks5Server struct {
	listener   net.Listener
	address    string
	activeConn int64
	lastError  string
	stopChan   chan struct{}
}

func (s *OtherService) StartSocks5Proxy(cfg Socks5Config) (Socks5Status, error) {
	if cfg.Port == 0 {
		return Socks5Status{}, errors.New("port is required")
	}
	if cfg.ListenIP == "" {
		return Socks5Status{}, errors.New("listen IP is required")
	}
	if cfg.ListenIP == "127.0.0.1" || strings.ToLower(cfg.ListenIP) == "localhost" {
		return Socks5Status{}, errors.New("listen IP cannot be 127.0.0.1")
	}
	addr := fmt.Sprintf("%s:%d", cfg.ListenIP, cfg.Port)
	fmt.Println("start locking")
	s.mu.Lock()
	if s.socksServer != nil {
		_ = s.socksServer.Close()
		s.socksServer = nil
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		return Socks5Status{}, err
	}
	server := &socks5Server{
		listener: listener,
		address:  addr,
		stopChan: make(chan struct{}),
	}
	s.socksServer = server
	status := s.currentSocksStatusLocked()
	s.mu.Unlock()
	go server.serve()
	return status, nil
}

func (s *OtherService) StopSocks5Proxy() (Socks5Status, error) {
	s.mu.Lock()
	if s.socksServer != nil {
		_ = s.socksServer.Close()
		s.socksServer = nil
	}
	status := s.currentSocksStatusLocked()
	s.mu.Unlock()
	return status, nil
}

func (s *OtherService) Socks5Status() Socks5Status {
	s.mu.Lock()
	status := s.currentSocksStatusLocked()
	s.mu.Unlock()
	return status
}

func (s *OtherService) currentSocksStatusLocked() Socks5Status {
	if s.socksServer == nil {
		return Socks5Status{}
	}
	return Socks5Status{
		Running:            true,
		Address:            s.socksServer.address,
		ActiveConnections:  atomic.LoadInt64(&s.socksServer.activeConn),
		Error:              s.socksServer.lastError,
		LastControlMessage: time.Now().Format(time.RFC3339),
	}
}

func (s *socks5Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			s.lastError = err.Error()
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *socks5Server) handleConnection(conn net.Conn) {
	atomic.AddInt64(&s.activeConn, 1)
	defer func() {
		atomic.AddInt64(&s.activeConn, -1)
		_ = conn.Close()
	}()

	buf := make([]byte, 1024)
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return
	}
	nMethods := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return
	}
	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return
	}
	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return
	}
	if buf[1] != 0x01 {
		_ = conn.Close()
		return
	}
	var addr string
	switch buf[3] {
	case 0x01: // IPv4
		if _, err := io.ReadFull(conn, buf[:4]); err != nil {
			return
		}
		ip := net.IPv4(buf[0], buf[1], buf[2], buf[3]).String()
		if _, err := io.ReadFull(conn, buf[:2]); err != nil {
			return
		}
		port := binary.BigEndian.Uint16(buf[:2])
		addr = fmt.Sprintf("%s:%d", ip, port)
	case 0x03: // Domain
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return
		}
		domainLen := int(buf[0])
		if _, err := io.ReadFull(conn, buf[:domainLen]); err != nil {
			return
		}
		domain := string(buf[:domainLen])
		if _, err := io.ReadFull(conn, buf[:2]); err != nil {
			return
		}
		port := binary.BigEndian.Uint16(buf[:2])
		addr = fmt.Sprintf("%s:%d", domain, port)
	default:
		return
	}
	target, err := net.Dial("tcp", addr)
	if err != nil {
		_, _ = conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer target.Close()
	_, _ = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	go io.Copy(target, conn)
	_, _ = io.Copy(conn, target)
}

func (s *socks5Server) Close() error {
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// GM SSL

type GMSSLServerConfig struct {
	ListenIP   string `json:"listenIp"`
	Port       int    `json:"port"`
	SignCertID string `json:"signCertId"`
	SignKeyID  string `json:"signKeyId"`
	EncCertID  string `json:"encCertId"`
	EncKeyID   string `json:"encKeyId"`
	ClientAuth bool   `json:"clientAuth"`
}

type GMSSLServerStatus struct {
	Running   bool   `json:"running"`
	Address   string `json:"address"`
	Error     string `json:"error"`
	StartedAt string `json:"startedAt"`
}

type gmsslServer struct {
	listener net.Listener
	address  string
	err      string
	started  time.Time
	stop     chan struct{}
}

type GMSSLClientConfig struct {
	ServerIP         string `json:"serverIp"`
	Port             int    `json:"port"`
	EnableClientAuth bool   `json:"enableClientAuth"`
	SignCertID       string `json:"signCertId"`
	SignKeyID        string `json:"signKeyId"`
	EncCertID        string `json:"encCertId"`
	EncKeyID         string `json:"encKeyId"`
	SkipVerify       bool   `json:"skipVerify"`
}

type GMSSLClientResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func (s *OtherService) StartGMSSLServer(cfg GMSSLServerConfig) (GMSSLServerStatus, error) {
	if cfg.ListenIP == "" || cfg.Port == 0 {
		return GMSSLServerStatus{}, errors.New("listen IP and port are required")
	}
	if cfg.SignCertID == "" || cfg.SignKeyID == "" || cfg.EncCertID == "" || cfg.EncKeyID == "" {
		return GMSSLServerStatus{}, errors.New("sign/encrypt certificates and keys are required")
	}

	signCert, err := s.loadTLCPIdentity(cfg.SignCertID, cfg.SignKeyID)
	if err != nil {
		return GMSSLServerStatus{}, err
	}
	encCert, err := s.loadTLCPIdentity(cfg.EncCertID, cfg.EncKeyID)
	if err != nil {
		return GMSSLServerStatus{}, err
	}

	tlcpConfig := &tlcp.Config{
		Certificates: []tlcp.Certificate{signCert, encCert},
		ClientAuth:   tlcp.NoClientCert,
	}
	if cfg.ClientAuth {
		tlcpConfig.ClientAuth = tlcp.RequireAndVerifyClientCert
	}

	addr := fmt.Sprintf("%s:%d", cfg.ListenIP, cfg.Port)
	listener, err := tlcp.Listen("tcp", addr, tlcpConfig)
	if err != nil {
		return GMSSLServerStatus{}, err
	}

	server := &gmsslServer{
		listener: listener,
		address:  addr,
		started:  time.Now(),
		stop:     make(chan struct{}),
	}

	s.mu.Lock()
	if s.gmServer != nil {
		_ = s.gmServer.Close()
	}
	s.gmServer = server
	status := s.currentGMStatusLocked()
	s.mu.Unlock()
	go server.serve()
	return status, nil
}

func (s *OtherService) StopGMSSLServer() (GMSSLServerStatus, error) {
	s.mu.Lock()
	if s.gmServer != nil {
		_ = s.gmServer.Close()
		s.gmServer = nil
	}
	status := s.currentGMStatusLocked()
	s.mu.Unlock()
	return status, nil
}

func (s *OtherService) GMSSLServerStatus() GMSSLServerStatus {
	s.mu.Lock()
	status := s.currentGMStatusLocked()
	s.mu.Unlock()
	return status
}

func (s *OtherService) currentGMStatusLocked() GMSSLServerStatus {
	if s.gmServer == nil {
		return GMSSLServerStatus{}
	}
	return GMSSLServerStatus{
		Running:   true,
		Address:   s.gmServer.address,
		Error:     s.gmServer.err,
		StartedAt: s.gmServer.started.Format(time.RFC3339),
	}
}

func (s *gmsslServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			s.err = err.Error()
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			_, _ = c.Write([]byte("GMSSL OK"))
		}(conn)
	}
}

func (s *gmsslServer) Close() error {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *OtherService) RunGMSSLClientTest(cfg GMSSLClientConfig) (GMSSLClientResult, error) {
	if cfg.ServerIP == "" || cfg.Port == 0 {
		return GMSSLClientResult{}, errors.New("server IP and port required")
	}
	addr := fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.Port)

	tlcpConfig := &tlcp.Config{
		InsecureSkipVerify: cfg.SkipVerify,
	}
	if cfg.EnableClientAuth {
		signCert, err := s.loadTLCPIdentity(cfg.SignCertID, cfg.SignKeyID)
		if err != nil {
			return GMSSLClientResult{}, err
		}
		encCert, err := s.loadTLCPIdentity(cfg.EncCertID, cfg.EncKeyID)
		if err != nil {
			return GMSSLClientResult{}, err
		}
		tlcpConfig.Certificates = []tlcp.Certificate{signCert, encCert}
	}

	conn, err := tlcp.Dial("tcp", addr, tlcpConfig)
	if err != nil {
		return GMSSLClientResult{
			Success:   false,
			Message:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}
	defer conn.Close()

	buff := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, readErr := conn.Read(buff)
	msg := "handshake completed"
	if readErr == nil && n > 0 {
		msg = fmt.Sprintf("received: %s", string(buff[:n]))
	}
	return GMSSLClientResult{
		Success:   true,
		Message:   msg,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *OtherService) loadTLCPIdentity(certID, keyID string) (tlcp.Certificate, error) {
	certExport, err := s.crypto.ExportCertificate(certID)
	if err != nil {
		return tlcp.Certificate{}, err
	}
	key, err := s.crypto.ExportStoredKey(keyID)
	if err != nil {
		return tlcp.Certificate{}, err
	}
	if key.PrivatePEM == "" {
		return tlcp.Certificate{}, errors.New("selected key does not contain private PEM")
	}
	return tlcp.X509KeyPair([]byte(certExport.Cert.CertPEM), []byte(key.PrivatePEM))
}
