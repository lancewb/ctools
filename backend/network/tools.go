package network

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DNSLookupRequest struct {
	Host string `json:"host"`
}

type DNSLookupResult struct {
	Host      string   `json:"host"`
	CNAME     string   `json:"cname"`
	Addresses []string `json:"addresses"`
	MX        []string `json:"mx"`
	NS        []string `json:"ns"`
	TXT       []string `json:"txt"`
	Error     string   `json:"error"`
}

type PortScanRequest struct {
	Host          string `json:"host"`
	Ports         string `json:"ports"`
	TimeoutMillis int    `json:"timeoutMillis"`
}

type PortScanResult struct {
	Port          int    `json:"port"`
	Open          bool   `json:"open"`
	LatencyMillis int64  `json:"latencyMillis"`
	Error         string `json:"error"`
}

type TCPClientRequest struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Payload       string `json:"payload"`
	PayloadFormat string `json:"payloadFormat"`
	TimeoutMillis int    `json:"timeoutMillis"`
}

type TCPClientResult struct {
	Connected      bool   `json:"connected"`
	LatencyMillis  int64  `json:"latencyMillis"`
	Response       string `json:"response"`
	ResponseBase64 string `json:"responseBase64"`
	Error          string `json:"error"`
}

type PrometheusScrapeRequest struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Scheme        string `json:"scheme"`
	Path          string `json:"path"`
	TimeoutMillis int    `json:"timeoutMillis"`
}

type PrometheusMetric struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	Raw    string            `json:"raw"`
}

type PrometheusScrapeResult struct {
	URL       string             `json:"url"`
	Timestamp string             `json:"timestamp"`
	Metrics   []PrometheusMetric `json:"metrics"`
	Error     string             `json:"error"`
}

func (n *NetworkService) LookupDNS(req DNSLookupRequest) DNSLookupResult {
	host := normalizeHost(req.Host)
	result := DNSLookupResult{Host: host}
	if host == "" {
		result.Error = "host is required"
		return result
	}
	if cname, err := net.LookupCNAME(host); err == nil {
		result.CNAME = strings.TrimSuffix(cname, ".")
	}
	if ips, err := net.LookupIP(host); err == nil {
		for _, ip := range ips {
			result.Addresses = append(result.Addresses, ip.String())
		}
		sort.Strings(result.Addresses)
	} else {
		result.Error = err.Error()
	}
	if mxs, err := net.LookupMX(host); err == nil {
		for _, mx := range mxs {
			result.MX = append(result.MX, fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, ".")))
		}
	}
	if nss, err := net.LookupNS(host); err == nil {
		for _, ns := range nss {
			result.NS = append(result.NS, strings.TrimSuffix(ns.Host, "."))
		}
	}
	if txts, err := net.LookupTXT(host); err == nil {
		result.TXT = append(result.TXT, txts...)
	}
	return result
}

func (n *NetworkService) ScanPorts(req PortScanRequest) []PortScanResult {
	host := normalizeHost(req.Host)
	ports, err := parsePortList(req.Ports)
	if host == "" || err != nil {
		msg := "host is required"
		if err != nil {
			msg = err.Error()
		}
		return []PortScanResult{{Error: msg}}
	}
	timeout := timeoutFromMillis(req.TimeoutMillis, 800*time.Millisecond)
	results := make([]PortScanResult, len(ports))
	sem := make(chan struct{}, 64)
	done := make(chan PortScanResult, len(ports))
	for _, port := range ports {
		sem <- struct{}{}
		go func(port int) {
			defer func() { <-sem }()
			start := time.Now()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
			latency := time.Since(start).Milliseconds()
			result := PortScanResult{Port: port, LatencyMillis: latency}
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Open = true
				_ = conn.Close()
			}
			done <- result
		}(port)
	}
	for i := range ports {
		results[i] = <-done
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Port < results[j].Port })
	return results
}

func (n *NetworkService) SendTCP(req TCPClientRequest) TCPClientResult {
	host := normalizeHost(req.Host)
	if host == "" || req.Port <= 0 || req.Port > 65535 {
		return TCPClientResult{Error: "host and port are required"}
	}
	payload, err := decodeTCPPayload(req.Payload, req.PayloadFormat)
	if err != nil {
		return TCPClientResult{Error: err.Error()}
	}
	timeout := timeoutFromMillis(req.TimeoutMillis, 3*time.Second)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(req.Port)), timeout)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return TCPClientResult{LatencyMillis: latency, Error: err.Error()}
	}
	defer conn.Close()
	result := TCPClientResult{Connected: true, LatencyMillis: latency}
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if len(payload) > 0 {
		if _, err := conn.Write(payload); err != nil {
			result.Error = err.Error()
			return result
		}
	}
	var buf bytes.Buffer
	_, _ = io.CopyN(&buf, conn, 64*1024)
	if buf.Len() > 0 {
		result.Response = buf.String()
		result.ResponseBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
	}
	return result
}

func (n *NetworkService) ScrapePrometheus(req PrometheusScrapeRequest) PrometheusScrapeResult {
	targetURL, err := buildPrometheusURL(req)
	if err != nil {
		return PrometheusScrapeResult{Error: err.Error(), Timestamp: time.Now().Format(time.RFC3339)}
	}
	timeout := timeoutFromMillis(req.TimeoutMillis, 5*time.Second)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(targetURL)
	if err != nil {
		return PrometheusScrapeResult{URL: targetURL, Timestamp: time.Now().Format(time.RFC3339), Error: err.Error()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return PrometheusScrapeResult{URL: targetURL, Timestamp: time.Now().Format(time.RFC3339), Error: err.Error()}
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return PrometheusScrapeResult{URL: targetURL, Timestamp: time.Now().Format(time.RFC3339), Error: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
	}
	metrics := parsePrometheusMetrics(body)
	return PrometheusScrapeResult{
		URL:       targetURL,
		Timestamp: time.Now().Format(time.RFC3339),
		Metrics:   metrics,
	}
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.Trim(host, " /")
}

func parsePortList(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("ports are required")
	}
	seen := map[int]bool{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid port range: %s", part)
			}
			if start > end {
				start, end = end, start
			}
			for port := start; port <= end; port++ {
				if port > 0 && port <= 65535 {
					seen[port] = true
				}
			}
			continue
		}
		port, err := strconv.Atoi(part)
		if err != nil || port <= 0 || port > 65535 {
			return nil, fmt.Errorf("invalid port: %s", part)
		}
		seen[port] = true
	}
	ports := make([]int, 0, len(seen))
	for port := range seen {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	if len(ports) == 0 {
		return nil, errors.New("no valid ports")
	}
	if len(ports) > 1024 {
		return nil, errors.New("scan is limited to 1024 ports")
	}
	return ports, nil
}

func decodeTCPPayload(payload, format string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "base64", "b64":
		return base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	case "hex":
		normalized := strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(payload)
		return hex.DecodeString(normalized)
	default:
		return []byte(payload), nil
	}
}

func timeoutFromMillis(value int, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return time.Duration(value) * time.Millisecond
}

func buildPrometheusURL(req PrometheusScrapeRequest) (string, error) {
	host := normalizeHost(req.Host)
	if host == "" || req.Port <= 0 || req.Port > 65535 {
		return "", errors.New("host and port are required")
	}
	scheme := strings.ToLower(strings.TrimSpace(req.Scheme))
	if scheme == "" {
		scheme = "http"
	}
	if scheme != "http" && scheme != "https" {
		return "", errors.New("scheme must be http or https")
	}
	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(req.Port)),
		Path:   path,
	}
	return u.String(), nil
}

func parsePrometheusMetrics(data []byte) []PrometheusMetric {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	metrics := []PrometheusMetric{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			continue
		}
		name, labels := splitPrometheusNameLabels(fields[0])
		if name == "" {
			continue
		}
		metrics = append(metrics, PrometheusMetric{
			Name:   name,
			Labels: labels,
			Value:  value,
			Raw:    line,
		})
	}
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Name == metrics[j].Name {
			return metrics[i].Raw < metrics[j].Raw
		}
		return metrics[i].Name < metrics[j].Name
	})
	return metrics
}

func splitPrometheusNameLabels(token string) (string, map[string]string) {
	open := strings.IndexByte(token, '{')
	if open < 0 {
		return token, map[string]string{}
	}
	close := strings.LastIndexByte(token, '}')
	if close < open {
		return "", nil
	}
	name := token[:open]
	labels := map[string]string{}
	for _, part := range splitLabelPairs(token[open+1 : close]) {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		value = strings.Trim(value, `"`)
		labels[strings.TrimSpace(key)] = value
	}
	return name, labels
}

func splitLabelPairs(raw string) []string {
	parts := []string{}
	var current strings.Builder
	escaped := false
	inQuote := false
	for _, r := range raw {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			current.WriteRune(r)
			escaped = true
		case r == '"':
			current.WriteRune(r)
			inQuote = !inQuote
		case r == ',' && !inQuote:
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}
