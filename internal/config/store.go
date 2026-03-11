package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Store 配置存储
type Store struct {
	path string
}

// NewStore 创建配置存储
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Load 加载配置，如果文件不存在则返回默认配置
func (s *Store) Load() (*Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 填充默认值
	if cfg.ProxyPort == 0 {
		cfg.ProxyPort = 443
	}
	if cfg.AdminPort == 0 {
		cfg.AdminPort = 8442
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}

	return cfg, nil
}

// Save 保存配置
func (s *Store) Save(cfg *Config) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// Path 返回配置文件路径
func (s *Store) Path() string {
	return s.path
}