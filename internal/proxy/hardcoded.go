package proxy

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// HardcodedEndpoint 硬编码端点处理
// 这些端点在 Claude Code 二进制中硬编码指向 api.anthropic.com
// 需要返回正确的响应格式，避免客户端报错或长时间等待

// isHardcodedEndpoint 检查是否为硬编码端点
func isHardcodedEndpoint(path string) bool {
	// 精确匹配的端点
	exactMatches := []string{
		"/v1/me",
		"/api/claude_cli_feedback",
		"/api/claude_code_shared_session_transcripts",
		"/api/oauth/claude_cli/create_api_key",
		"/api/oauth/claude_cli/role",
		"/api/claude_code/organizations/metrics_enabled",
	}

	for _, match := range exactMatches {
		if path == match {
			return true
		}
	}

	// 前缀匹配的端点
	prefixMatches := []string{
		"/api/claude_code/metric",
		"/api/claude_code/organization",
		"/api/web/domain_info",
	}

	for _, prefix := range prefixMatches {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// handleHardcodedEndpoint 处理硬编码端点请求
func (h *Handler) handleHardcodedEndpoint(w http.ResponseWriter, r *http.Request) bool {
	path := r.URL.Path

	// 快速检查是否为硬编码端点
	if !isHardcodedEndpoint(path) {
		return false
	}

	// 消耗请求体以确保连接可复用
	drainRequestBody(r)

	switch {
	// 反馈提交 - POST /api/claude_cli_feedback
	case path == "/api/claude_cli_feedback":
		h.handleFeedback(w, r)
		return true

	// 指标上报 - POST /api/claude_code/metric
	case strings.HasPrefix(path, "/api/claude_code/metric"):
		h.handleMetric(w, r)
		return true

	// 组织指标开关 - GET /api/claude_code/organizations/metrics_enabled
	case path == "/api/claude_code/organizations/metrics_enabled":
		h.handleMetricsEnabled(w, r)
		return true

	// 组织信息 - GET /api/claude_code/organization
	case strings.HasPrefix(path, "/api/claude_code/organization"):
		h.handleOrganization(w, r)
		return true

	// 会话记录共享 - POST /api/claude_code_shared_session_transcripts
	case path == "/api/claude_code_shared_session_transcripts":
		h.handleSessionTranscripts(w, r)
		return true

	// 域名信息 - GET /api/web/domain_info?domain=xxx
	case strings.HasPrefix(path, "/api/web/domain_info"):
		h.handleDomainInfo(w, r)
		return true

	// 创建 API 密钥 - POST /api/oauth/claude_cli/create_api_key
	case path == "/api/oauth/claude_cli/create_api_key":
		h.handleCreateAPIKey(w, r)
		return true

	// 角色信息 - GET /api/oauth/claude_cli/role
	case path == "/api/oauth/claude_cli/role":
		h.handleRole(w, r)
		return true

	// 用户信息 - GET /v1/me
	case path == "/v1/me":
		h.handleMe(w, r)
		return true
	}

	return false
}

// handleFeedback 处理反馈提交
// 响应格式: { "feedback_id": "xxx" }
func (h *Handler) handleFeedback(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling feedback request: %s", r.URL.Path)

	response := map[string]interface{}{
		"feedback_id": generateID(),
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleMetric 处理指标上报
// 响应格式: { "success": true }
func (h *Handler) handleMetric(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling metric request: %s", r.URL.Path)

	response := map[string]any{
		"success": true,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleMetricsEnabled 处理组织指标开关请求
// 响应格式: { "metrics_enabled": false }
func (h *Handler) handleMetricsEnabled(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling metrics_enabled request: %s", r.URL.Path)

	response := map[string]any{
		"metrics_enabled": false,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleOrganization 处理组织信息请求
// 响应格式: { "metrics_enabled": false } 或空对象
func (h *Handler) handleOrganization(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling organization request: %s", r.URL.Path)

	// 默认组织信息响应
	response := map[string]any{
		"organization_id":  "local-proxy",
		"metrics_enabled":  false,
		"can_use_otel":     false,
		"has_subscription": false,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleSessionTranscripts 处理会话记录共享
// 响应格式: { "success": true, "transcript_id": "xxx" }
func (h *Handler) handleSessionTranscripts(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling session transcripts request: %s", r.URL.Path)

	response := map[string]interface{}{
		"success":        true,
		"transcript_id":  generateID(),
		"share_url":      "",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleDomainInfo 处理域名信息请求
// 响应格式: { "can_fetch": true }
func (h *Handler) handleDomainInfo(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	log.Printf("[Hardcoded] Handling domain info request: domain=%s", domain)

	// 默认允许所有域名
	response := map[string]interface{}{
		"can_fetch": true,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleCreateAPIKey 处理创建 API 密钥请求
// 响应格式: { "api_key": "xxx", "created_at": "..." }
func (h *Handler) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling create API key request: %s", r.URL.Path)

	response := map[string]interface{}{
		"api_key":    "sk-ant-api03-local-proxy-" + generateID(),
		"created_at": "2026-03-11T00:00:00Z",
		"type":       "api_key",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleRole 处理角色信息请求
// 响应格式: { "role": "user", "permissions": [] }
func (h *Handler) handleRole(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling role request: %s", r.URL.Path)

	response := map[string]interface{}{
		"role":        "user",
		"permissions": []string{},
		"can_upgrade": false,
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// handleMe 处理用户信息请求
// 响应格式: { "id": "xxx", "type": "user", ... }
func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Hardcoded] Handling /v1/me request: %s", r.URL.Path)

	response := map[string]interface{}{
		"id":           "user-local-proxy",
		"type":         "user",
		"email":        "local@proxy.dev",
		"display_name": "Local Proxy User",
		"created_at":   "2026-01-01T00:00:00Z",
		"organization": map[string]interface{}{
			"id":   "org-local-proxy",
			"name": "Local Proxy Organization",
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse 写入 JSON 响应
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// generateID 生成唯一的 ID 字符串
// 注意：这不是标准 UUID，仅用于生成唯一标识符
func generateID() string {
	return "proxy-" + randomHex(8) + "-" + randomHex(4) + "-" + randomHex(4) + "-" + randomHex(4) + "-" + randomHex(12)
}

// randomHex 生成指定长度的十六进制字符串
func randomHex(length int) string {
	b := make([]byte, (length+1)/2)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return strings.Repeat("0", length)
	}
	return hex.EncodeToString(b)[:length]
}

// drainRequestBody 消耗并关闭请求体，确保 HTTP 连接可复用
func drainRequestBody(r *http.Request) {
	if r.Body != nil {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			log.Printf("Warning: failed to drain request body: %v", err)
		}
		r.Body.Close()
	}
}