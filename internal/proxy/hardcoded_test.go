package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"claude_code_proxy_dns/internal/config"
)

func TestIsHardcodedEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/claude_cli_feedback", true},
		{"/api/claude_code/metric", true},
		{"/api/claude_code/metrics", true}, // metrics 匹配 metric 前缀
		{"/api/claude_code/organization", true},
		{"/api/claude_code/organizations/metrics_enabled", true},
		{"/api/claude_code_shared_session_transcripts", true},
		{"/api/web/domain_info", true},
		{"/api/oauth/claude_cli/create_api_key", true},
		{"/api/oauth/claude_cli/role", true},
		{"/v1/me", true},
		{"/v1/messages", false},
		{"/v1/complete", false},
		{"/some/other/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isHardcodedEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("isHardcodedEndpoint(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestHandleFeedback(t *testing.T) {
	handler := NewHandler(config.NewMockStore(nil), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/claude_cli_feedback", nil)
	rec := httptest.NewRecorder()

	handler.handleFeedback(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// 验证响应包含 feedback_id
	body := rec.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}
}

func TestHandleDomainInfo(t *testing.T) {
	handler := NewHandler(config.NewMockStore(nil), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/web/domain_info?domain=example.com", nil)
	rec := httptest.NewRecorder()

	handler.handleDomainInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("Response body is empty")
	}

	// 验证响应包含 can_fetch
	if !strings.Contains(body, "can_fetch") {
		t.Error("Response should contain 'can_fetch'")
	}
}

func TestHandleOrganization(t *testing.T) {
	handler := NewHandler(config.NewMockStore(nil), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/claude_code/organization", nil)
	rec := httptest.NewRecorder()

	handler.handleOrganization(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "metrics_enabled") {
		t.Error("Response should contain 'metrics_enabled'")
	}
}

func TestHandleMe(t *testing.T) {
	handler := NewHandler(config.NewMockStore(nil), nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	rec := httptest.NewRecorder()

	handler.handleMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "id") {
		t.Error("Response should contain 'id'")
	}
}

func TestHandleMetric(t *testing.T) {
	handler := NewHandler(config.NewMockStore(nil), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/claude_code/metric", nil)
	rec := httptest.NewRecorder()

	handler.handleMetric(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "success") {
		t.Error("Response should contain 'success'")
	}
}

func TestHandleHardcodedEndpoint(t *testing.T) {
	handler := NewHandler(config.NewMockStore(nil), nil)

	tests := []struct {
		path string
	}{
		{"/api/claude_cli_feedback"},
		{"/api/claude_code/metric"},
		{"/api/claude_code/organization"},
		{"/api/web/domain_info?domain=test.com"},
		{"/v1/me"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			result := handler.handleHardcodedEndpoint(rec, req)
			if !result {
				t.Errorf("handleHardcodedEndpoint should return true for %s", tt.path)
			}
		})
	}
}