package admin

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Auth 认证管理器
type Auth struct {
	passwordHash    []byte
	sessions        map[string]time.Time
	sessionDuration time.Duration
	mu              sync.RWMutex

	// 登录保护
	failedAttempts  int
	maxAttempts     int
	lockDuration    time.Duration
	lockedUntil     time.Time

	// Session 清理
	stopCleanup chan struct{}
}

// NewAuth 创建认证管理器
func NewAuth(password string) *Auth {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	auth := &Auth{
		passwordHash:    hash,
		sessions:        make(map[string]time.Time),
		sessionDuration: 24 * time.Hour,
		maxAttempts:     5,
		lockDuration:    5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}
	go auth.cleanupExpiredSessions()
	return auth
}

// NewAuthWithConfig 创建带配置的认证管理器
func NewAuthWithConfig(password string, maxAttempts int, lockDuration time.Duration) *Auth {
	auth := NewAuth(password)
	auth.maxAttempts = maxAttempts
	auth.lockDuration = lockDuration
	return auth
}

// VerifyPassword 验证密码
func (a *Auth) VerifyPassword(password string) bool {
	// 检查是否被锁定
	if a.IsLocked() {
		return false
	}

	err := bcrypt.CompareHashAndPassword(a.passwordHash, []byte(password))
	return err == nil
}

// GenerateToken 生成会话 token
func (a *Auth) GenerateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)

	a.mu.Lock()
	a.sessions[token] = time.Now().Add(a.sessionDuration)
	a.mu.Unlock()

	return token
}

// ValidateToken 验证 token
func (a *Auth) ValidateToken(token string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	expiry, exists := a.sessions[token]
	if !exists {
		return false
	}

	return time.Now().Before(expiry)
}

// InvalidateToken 使 token 失效
func (a *Auth) InvalidateToken(token string) {
	a.mu.Lock()
	delete(a.sessions, token)
	a.mu.Unlock()
}

// RecordFailedAttempt 记录失败尝试
func (a *Auth) RecordFailedAttempt() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.failedAttempts++
	if a.failedAttempts >= a.maxAttempts {
		a.lockedUntil = time.Now().Add(a.lockDuration)
	}
}

// ResetAttempts 重置失败计数
func (a *Auth) ResetAttempts() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.failedAttempts = 0
	a.lockedUntil = time.Time{}
}

// IsLocked 检查是否被锁定
func (a *Auth) IsLocked() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.lockedUntil.IsZero() {
		return false
	}

	return time.Now().Before(a.lockedUntil)
}

// SetPassword 设置密码
func (a *Auth) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.passwordHash = hash
	a.mu.Unlock()

	return nil
}

// Close 停止清理 goroutine
func (a *Auth) Close() {
	close(a.stopCleanup)
}

// cleanupExpiredSessions 定期清理过期 session
func (a *Auth) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			now := time.Now()
			for token, expiry := range a.sessions {
				if now.After(expiry) {
					delete(a.sessions, token)
				}
			}
			a.mu.Unlock()
		case <-a.stopCleanup:
			return
		}
	}
}