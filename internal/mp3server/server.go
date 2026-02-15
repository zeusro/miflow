// Package mp3server provides HTTP file server that maps local paths to accessible URLs.
// Mapping: /Users/zeusro/Music/QQ音乐/Taylor Swift-Red.flac -> http://本机ip:端口/Users/zeusro/Music/QQ音乐/Taylor%20Swift-Red.flac
package mp3server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds server configuration.
type Config struct {
	Addr       string // e.g. ":8090"
	Host       string // 本机 IP，空则自动检测
	LogRequest bool   // 打印每个 HTTP 请求
}

// Server runs an HTTP file server and provides path-to-URL mapping.
type Server struct {
	cfg   Config
	ln    net.Listener
	mux   *http.ServeMux
	host  string
	port  string
	root  string
	ready chan struct{}
}

// New creates a new Server. root is the filesystem root for serving (use "/" for full path mapping).
func New(cfg Config, root string) (*Server, error) {
	if root == "" {
		root = "/"
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	// For full path mapping, we need root="/"
	if root != "/" {
		root = filepath.Clean(root)
	}
	s := &Server{
		cfg:   cfg,
		mux:   http.NewServeMux(),
		root:  root,
		ready: make(chan struct{}),
	}
	fs := http.FileServer(http.Dir(root))
	if s.cfg.LogRequest {
		s.mux.Handle("/", logRequestHandler(fs, root))
	} else {
		s.mux.Handle("/", fs)
	}
	return s, nil
}

// logRequestHandler 包装 handler，打印请求路径及映射到的本地文件路径。
func logRequestHandler(h http.Handler, root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" {
			path = "/"
		}
		// r.URL.Path 已被 Go 解码，直接对应本地路径（root="/" 时）
		localPath := path
		if root != "/" && root != "" {
			localPath = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(path, "/")))
		}
		log.Printf("[请求] %s %s -> 本地: %s", r.Method, path, localPath)
		rw := &statusRecorder{ResponseWriter: w, status: 200}
		h.ServeHTTP(rw, r)
		log.Printf("[响应] %s %s -> %d", r.Method, path, rw.status)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Start starts the HTTP server. Call PathToURL to get the URL for a file.
func (s *Server) Start() error {
	addr := s.cfg.Addr
	if addr == "" {
		addr = ":8090"
	}
	if port := parsePort(addr); port != "" {
		killProcessOnPort(port)
		time.Sleep(200 * time.Millisecond)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("启动 HTTP 服务失败: %w", err)
	}
	s.ln = ln
	listenAddr := ln.Addr().String()
	s.host = s.cfg.Host
	if s.host == "" {
		s.host = getListenHost(listenAddr)
	}
	_, s.port, _ = net.SplitHostPort(listenAddr)
	if s.port == "" {
		s.port = "8090"
	}
	go func() {
		close(s.ready)
		if err := http.Serve(ln, s.mux); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			// log if needed
		}
	}()
	return nil
}

// Close stops the server.
func (s *Server) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

// WaitReady blocks until the server is ready or timeout. 需先调用 Start()。
func (s *Server) WaitReady(timeout time.Duration) bool {
	select {
	case <-s.ready:
		return waitPortReady("127.0.0.1", s.port, timeout)
	case <-time.After(timeout):
		return false
	}
}

// WaitPortReady 探测端口是否可连接，不依赖 Start()。用于 mp3 单独启动时验证服务就绪。
func (s *Server) WaitPortReady(timeout time.Duration) bool {
	s.ResolveHostPort()
	return waitPortReady("127.0.0.1", s.port, timeout)
}

// ResolveHostPort 从 Config 解析 host/port，无需启动服务。用于 mp3 单独启动时，xiaomusic 仅需生成 URL。
func (s *Server) ResolveHostPort() {
	if s.host != "" && s.port != "" {
		return
	}
	addr := s.cfg.Addr
	if addr == "" {
		addr = ":8090"
	}
	s.port = parsePort(addr)
	if s.port == "" {
		s.port = "8090"
	}
	s.host = s.cfg.Host
	if s.host == "" {
		s.host = getListenHost(":" + s.port)
	}
}

// PathToURL converts a local file path to an accessible HTTP URL.
// 不依赖 Start()：若 host/port 未设置则从 Config 解析，支持 mp3 单独启动的场景。
// Mapping: /Users/zeusro/Music/QQ音乐/Taylor Swift-Red.flac -> http://host:port/Users/zeusro/Music/QQ%E9%9F%B3%E4%B9%90/Taylor%20Swift-Red.flac
func (s *Server) PathToURL(target string) (string, error) {
	s.ResolveHostPort()
	target, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(target); err != nil {
		return "", fmt.Errorf("文件不存在: %s (%w)", target, err)
	}
	var rel string
	if s.root == "/" || s.root == "" {
		// Full path mapping: use path as-is (leading /)
		rel = target
		if !strings.HasPrefix(rel, "/") {
			rel = "/" + rel
		}
	} else {
		rel, err = filepath.Rel(s.root, target)
		if err != nil {
			return "", err
		}
		rel = "/" + strings.ReplaceAll(rel, string(filepath.Separator), "/")
	}
	// URL encode each path segment
	parts := strings.Split(strings.Trim(rel, "/"), "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	pathEnc := "/" + strings.Join(parts, "/")
	return fmt.Sprintf("http://%s:%s%s", s.host, s.port, pathEnc), nil
}

// Host returns the host used in URLs.
func (s *Server) Host() string { return s.host }

// Port returns the port.
func (s *Server) Port() string { return s.port }

func parsePort(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	if port == "" || port == "0" {
		return ""
	}
	return port
}

func killProcessOnPort(port string) {
	cmd := exec.Command("lsof", "-i", ":"+port, "-t")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	pids := strings.Fields(strings.TrimSpace(string(out)))
	for _, pid := range pids {
		if pid == "" {
			continue
		}
		if strconv.Itoa(os.Getpid()) == pid {
			continue
		}
		_ = exec.Command("kill", pid).Run()
	}
}

func waitPortReady(host, port string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort(host, port)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func getListenHost(listenAddr string) string {
	host, _, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return "127.0.0.1"
	}
	if host != "" && host != "0.0.0.0" && host != "::" && !strings.HasPrefix(host, "[::]") {
		return host
	}
	if ip := getOutboundIP(); ip != "" {
		return ip
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	var candidates []string
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
			continue
		}
		ipStr := ipnet.IP.String()
		if strings.HasPrefix(ipStr, "172.17.") || strings.HasPrefix(ipStr, "172.18.") ||
			strings.HasPrefix(ipStr, "192.168.65.") {
			continue
		}
		candidates = append(candidates, ipStr)
	}
	for _, ip := range candidates {
		if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
			return ip
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return "127.0.0.1"
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP == nil {
		return ""
	}
	ip := addr.IP.String()
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return ip
	}
	return ""
}
