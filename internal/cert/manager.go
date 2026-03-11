package cert

import (
	"errors"
)

// 错误定义
var (
	ErrInvalidPEM      = errors.New("invalid PEM format")
	ErrCANotFound       = errors.New("CA certificate not found")
	ErrCertNotAfter     = errors.New("certificate has expired")
	ErrCertNotYetValid  = errors.New("certificate is not yet valid")
)

// Manager 证书管理器
type Manager struct {
	dataDir string
}

// NewManager 创建证书管理器
func NewManager(dataDir string) *Manager {
	return &Manager{dataDir: dataDir}
}