package admin

import (
	"encoding/json"
	"net/http"
	"time"
)

// handleLogin 处理登录
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 检查是否被锁定
	if s.auth.IsLocked() {
		time.Sleep(1 * time.Second) // 延迟响应
		http.Error(w, `{"error": "account locked"}`, http.StatusTooManyRequests)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	if !s.auth.VerifyPassword(req.Password) {
		s.auth.RecordFailedAttempt()
		time.Sleep(1 * time.Second) // 延迟响应
		http.Error(w, `{"error": "invalid password"}`, http.StatusUnauthorized)
		return
	}

	// 重置失败计数
	s.auth.ResetAttempts()

	// 生成 session token
	token := s.auth.GenerateToken()

	// 设置 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleLogout 处理登出
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		s.auth.InvalidateToken(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleConfig 处理配置请求
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getConfig(w, r)
	case http.MethodPut:
		s.updateConfig(w, r)
	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// getConfig 获取配置
func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	// TODO: 从配置管理器获取
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"backend_url": "https://open.bigmodel.cn/api/anthropic",
	})
}

// updateConfig 更新配置
func (s *Server) updateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BackendURL string `json:"backend_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	// TODO: 更新配置管理器

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleStatus 处理状态请求
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running":        true,
		"backend_url":    "https://open.bigmodel.cn/api/anthropic",
		"uptime":         time.Since(s.startTime).String(),
		"requests_total": 0,
	})
}

// handleCertificates 处理证书信息请求
func (s *Server) handleCertificates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ca_cert_path":      "./data/ca.crt",
		"server_cert_path":  "./data/server.crt",
		"ca_expires_at":     time.Now().AddDate(10, 0, 0),
		"server_expires_at": time.Now().AddDate(10, 0, 0),
	})
}

// handleTestBackend 测试后端连接
func (s *Server) handleTestBackend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BackendURL string `json:"backend_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request"}`, http.StatusBadRequest)
		return
	}

	// 测试连接
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(req.BackendURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"status_code":  resp.StatusCode,
	})
}

// SetConfigManager 设置配置管理器
func (s *Server) SetConfigManager(cm interface{}) {
	// TODO: 实现配置管理器集成
}