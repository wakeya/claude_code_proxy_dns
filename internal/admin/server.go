package admin

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"claude_code_proxy_dns/internal/config"
)

// StatsProvider 统计数据提供者接口
type StatsProvider interface {
	Stats() (total int64, last time.Time, uptime time.Duration)
}

// Server 配置服务
type Server struct {
	config        *AdminConfig
	auth          *Auth
	server        *http.Server
	startTime     time.Time
	configStore   config.ConfigStore
	statsProvider StatsProvider
}

// AdminConfig 配置服务配置
type AdminConfig struct {
	Password   string
	CertFile   string
	KeyFile    string
	ConfigPath string
}

// NewServer 创建配置服务
func NewServer(cfg *AdminConfig, configStore config.ConfigStore, statsProvider StatsProvider) *Server {
	return &Server{
		config:        cfg,
		auth:          NewAuth(cfg.Password),
		startTime:     time.Now(),
		configStore:   configStore,
		statsProvider: statsProvider,
	}
}

// Start 启动配置服务
func (s *Server) Start(addr string, frontendFS embed.FS) error {
	// 创建路由
	mux := http.NewServeMux()

	// 静态文件
	staticFS, _ := fs.Sub(frontendFS, "dist")
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/", s.authMiddleware(fileServer))

	// API 路由
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/api/config", s.authMiddlewareFunc(s.handleConfig))
	mux.HandleFunc("/api/status", s.authMiddlewareFunc(s.handleStatus))
	mux.HandleFunc("/api/certificates", s.authMiddlewareFunc(s.handleCertificates))
	mux.HandleFunc("/api/config/test", s.authMiddlewareFunc(s.handleTestBackend))

	// 供应商 API 路由
	mux.HandleFunc("/api/providers", s.authMiddlewareFunc(s.handleProviders))
	mux.HandleFunc("/api/providers/test", s.authMiddlewareFunc(s.handleTestProvider))
	mux.HandleFunc("/api/providers/", s.authMiddlewareFunc(s.handleProviderRoutes))

	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Admin server starting on %s", addr)
	return s.server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
}

// Stop 停止配置服务
func (s *Server) Stop(ctx context.Context) error {
	// 停止 session 清理 goroutine
	s.auth.Close()

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// authMiddleware 认证中间件（用于静态文件）
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 登录页面不需要认证
		if r.URL.Path == "/login.html" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		// API 路由单独处理
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			next.ServeHTTP(w, r)
			return
		}

		// 检查 session cookie
		cookie, err := r.Cookie("session")
		if err != nil || !s.auth.ValidateToken(cookie.Value) {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddlewareFunc 认证中间件（用于 API）
func (s *Server) authMiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || !s.auth.ValidateToken(cookie.Value) {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// GetAuth 获取认证管理器
func (s *Server) GetAuth() *Auth {
	return s.auth
}