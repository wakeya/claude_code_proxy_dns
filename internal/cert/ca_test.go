package cert

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateCA(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cert-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	// 生成 CA
	caCert, caKey, err := manager.GenerateCA()
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// 验证证书
	cert, err := x509.ParseCertificate(caCert)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	// 验证是 CA 证书
	if !cert.IsCA {
		t.Error("expected certificate to be CA")
	}

	// 验证有效期 (10年，考虑闰年可能多几天)
	validFor := cert.NotAfter.Sub(cert.NotBefore)
	expectedDuration := 10 * 365 * 24 * time.Hour
	tolerance := 72 * time.Hour // 允许 3 天容差（考虑闰年）

	if validFor < expectedDuration-tolerance || validFor > expectedDuration+tolerance {
		t.Errorf("expected validity ~10 years, got %v", validFor)
	}

	// 验证私钥
	if caKey == nil {
		t.Error("expected private key to be returned")
	}
}

func TestSaveAndLoadCA(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cert-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	// 生成并保存
	caCert, caKey, err := manager.GenerateCA()
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	if err := manager.SaveCA(caCert, caKey); err != nil {
		t.Fatalf("failed to save CA: %v", err)
	}

	// 加载
	loadedCert, loadedKey, err := manager.LoadCA()
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	// 验证证书相同
	parsed1, _ := x509.ParseCertificate(caCert)
	parsed2, _ := x509.ParseCertificate(loadedCert)

	if !parsed1.Equal(parsed2) {
		t.Error("loaded certificate does not match saved")
	}

	// 验证私钥
	if loadedKey == nil {
		t.Error("expected loaded private key to be returned")
	}

	// 验证文件存在
	if _, err := os.Stat(filepath.Join(tmpDir, "ca.crt")); err != nil {
		t.Error("ca.crt file not found")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "ca.key")); err != nil {
		t.Error("ca.key file not found")
	}
}