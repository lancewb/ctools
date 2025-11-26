package other

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	defaultPlantUMLServer  = "http://127.0.0.1:18080/plantuml"
	defaultPlantUMLTimeout = 12 * time.Second
)

var plantumlAcceptHeaders = map[string]string{
	"svg": "image/svg+xml",
	"png": "image/png",
	"txt": "text/plain;charset=utf-8",
}

// PlantUMLRenderRequest carries the parameters required to render a PlantUML diagram.
type PlantUMLRenderRequest struct {
	Source         string `json:"source"`
	Format         string `json:"format"` // svg, png, txt
	ServerURL      string `json:"serverUrl"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	UseBuiltin     bool   `json:"useBuiltin"`
}

// PlantUMLRenderResponse contains the rendered payload encoded as Base64.
type PlantUMLRenderResponse struct {
	MimeType   string `json:"mimeType"`
	Data       string `json:"data"`
	Bytes      int    `json:"bytes"`
	Generated  string `json:"generated"`
	ServerUsed string `json:"serverUsed"`
}

// RenderPlantUML sends a diagram to a PlantUML HTTP server and returns the rendered result.
func (s *OtherService) RenderPlantUML(req PlantUMLRenderRequest) (PlantUMLRenderResponse, error) {
	source := strings.TrimSpace(req.Source)
	if source == "" {
		return PlantUMLRenderResponse{}, errors.New("diagram source is required")
	}
	format := normalizePlantUMLFormat(req.Format)
	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = defaultPlantUMLTimeout
	}

	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if req.UseBuiltin {
		result, err := renderPlantUMLBuiltin(ctx, source, format)
		if err != nil {
			return PlantUMLRenderResponse{}, err
		}
		return result, nil
	}

	serverURL, err := normalizePlantUMLServer(req.ServerURL)
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}
	endpoint, err := buildPlantUMLEndpoint(serverURL, format)
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(source))
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if accept := plantumlAcceptHeaders[format]; accept != "" {
		httpReq.Header.Set("Accept", accept)
	}

	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return PlantUMLRenderResponse{}, fmt.Errorf("plantuml server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	mime := detectPlantUMLMime(format, resp.Header.Get("Content-Type"))
	return PlantUMLRenderResponse{
		MimeType:   mime,
		Data:       base64.StdEncoding.EncodeToString(body),
		Bytes:      len(body),
		Generated:  time.Now().Format(time.RFC3339),
		ServerUsed: serverURL,
	}, nil
}

func renderPlantUMLBuiltin(ctx context.Context, source, format string) (PlantUMLRenderResponse, error) {
	javaExec, err := findJavaExecutable(true)
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}
	jarPath, err := ensureEmbeddedPlantUMLJar()
	if err != nil {
		return PlantUMLRenderResponse{}, err
	}

	flag := "-tsvg"
	switch format {
	case "png":
		flag = "-tpng"
	case "txt":
		flag = "-ttxt"
	}

	args := []string{
		"-Dfile.encoding=UTF-8",
		"-Djava.awt.headless=true",
		"-jar",
		jarPath,
		flag,
		"-pipe",
	}
	cmd := exec.CommandContext(ctx, javaExec, args...)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Stdin = strings.NewReader(source)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return PlantUMLRenderResponse{}, fmt.Errorf("plantuml builtin error: %s", strings.TrimSpace(stderr.String()))
		}
		return PlantUMLRenderResponse{}, err
	}

	output := stdout.Bytes()
	return PlantUMLRenderResponse{
		MimeType:   detectPlantUMLMime(format, ""),
		Data:       base64.StdEncoding.EncodeToString(output),
		Bytes:      len(output),
		Generated:  time.Now().Format(time.RFC3339),
		ServerUsed: "builtin",
	}, nil
}

// CheckJavaRuntime verifies that a Java runtime is available and returns the version string.
func (s *OtherService) CheckJavaRuntime() (string, error) {
	path, err := findJavaExecutable(false)
	if err != nil {
		return "", errors.New("未检测到 Java 运行环境，请访问 https://adoptium.net/ 下载并安装 Java 11+")
	}
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "-version")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Java 执行失败: %w", err)
	}
	return strings.TrimSpace(output.String()), nil
}

func findJavaExecutable(preferGUI bool) (string, error) {
	if runtime.GOOS == "windows" {
		if preferGUI {
			if exe, err := exec.LookPath("javaw.exe"); err == nil {
				return exe, nil
			}
		}
		if exe, err := exec.LookPath("java.exe"); err == nil {
			return exe, nil
		}
	}
	return exec.LookPath("java")
}

func normalizePlantUMLFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "png":
		return "png"
	case "txt", "text":
		return "txt"
	default:
		return "svg"
	}
}

func normalizePlantUMLServer(raw string) (string, error) {
	server := strings.TrimSpace(raw)
	if server == "" {
		return defaultPlantUMLServer, nil
	}
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}
	parsed, err := url.Parse(server)
	if err != nil {
		return "", fmt.Errorf("invalid PlantUML server URL: %w", err)
	}
	if parsed.Host == "" {
		return "", errors.New("invalid PlantUML server URL: missing host")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String(), nil
}

func buildPlantUMLEndpoint(baseURL, format string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	combinedPath := path.Join(parsed.Path, format)
	if !strings.HasSuffix(combinedPath, "/") {
		combinedPath += "/"
	}
	parsed.Path = combinedPath
	return parsed.String(), nil
}

func detectPlantUMLMime(format, header string) string {
	if header != "" {
		return header
	}
	switch format {
	case "png":
		return "image/png"
	case "txt":
		return "text/plain;charset=utf-8"
	default:
		return "image/svg+xml"
	}
}
