// Package cert 提供证书管理功能
// Package cert provides certificate management functionality
package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hackmitm/pkg/logger"
)

// CertManager 证书管理器
// CertManager certificate manager
type CertManager struct {
	// caKey CA私钥
	caKey *ecdsa.PrivateKey
	// caCert CA证书
	caCert *x509.Certificate
	// certCache 证书缓存
	certCache map[string]*tls.Certificate
	// cacheMutex 缓存锁
	cacheMutex sync.RWMutex
	// certDir 证书存储目录
	certDir string
	// enableCache 启用缓存
	enableCache bool
	// cacheTTL 缓存TTL
	cacheTTL time.Duration
}

// CertOptions 证书选项
// CertOptions certificate options
type CertOptions struct {
	// CertDir 证书存储目录
	CertDir string
	// EnableCache 启用证书缓存
	EnableCache bool
	// CacheTTL 缓存TTL
	CacheTTL time.Duration
}

// NewCertManager 创建新的证书管理器
// NewCertManager creates a new certificate manager
func NewCertManager(opts CertOptions) (*CertManager, error) {
	if opts.CertDir == "" {
		opts.CertDir = "./certs"
	}
	if opts.CacheTTL == 0 {
		opts.CacheTTL = 24 * time.Hour
	}

	// 创建证书目录
	if err := os.MkdirAll(opts.CertDir, 0755); err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %w", err)
	}

	cm := &CertManager{
		certCache:   make(map[string]*tls.Certificate),
		certDir:     opts.CertDir,
		enableCache: opts.EnableCache,
		cacheTTL:    opts.CacheTTL,
	}

	// 初始化CA证书
	if err := cm.initCA(); err != nil {
		return nil, fmt.Errorf("初始化CA证书失败: %w", err)
	}

	logger.Info("证书管理器初始化成功")
	return cm, nil
}

// initCA 初始化CA证书
// initCA initializes CA certificate
func (cm *CertManager) initCA() error {
	caKeyPath := filepath.Join(cm.certDir, "ca-key.pem")
	caCertPath := filepath.Join(cm.certDir, "ca-cert.pem")

	// 检查CA文件是否存在
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		logger.Info("CA证书不存在，正在生成新的CA证书")
		return cm.generateCA()
	}

	// 加载现有CA证书
	return cm.loadCA(caKeyPath, caCertPath)
}

// generateCA 生成CA证书
// generateCA generates CA certificate
func (cm *CertManager) generateCA() error {
	// 生成ECDSA私钥
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("生成CA私钥失败: %w", err)
	}

	// 创建CA证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"HackMITM"},
			OrganizationalUnit: []string{"HackMITM Root CA"},
			CommonName:         "HackMITM Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10年有效期
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	// 自签名CA证书
	caCertBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("创建CA证书失败: %w", err)
	}

	// 解析CA证书
	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		return fmt.Errorf("解析CA证书失败: %w", err)
	}

	// 保存CA私钥
	caKeyPath := filepath.Join(cm.certDir, "ca-key.pem")
	keyBytes, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return fmt.Errorf("序列化CA私钥失败: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	if err := os.WriteFile(caKeyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("保存CA私钥失败: %w", err)
	}

	// 保存CA证书
	caCertPath := filepath.Join(cm.certDir, "ca-cert.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})

	if err := os.WriteFile(caCertPath, certPEM, 0644); err != nil {
		return fmt.Errorf("保存CA证书失败: %w", err)
	}

	cm.caKey = caKey
	cm.caCert = caCert

	logger.Info("CA证书生成成功")
	return nil
}

// loadCA 加载CA证书
// loadCA loads CA certificate
func (cm *CertManager) loadCA(keyPath, certPath string) error {
	// 加载CA私钥
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("读取CA私钥失败: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("解码CA私钥失败")
	}

	caKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析CA私钥失败: %w", err)
	}

	// 加载CA证书
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取CA证书失败: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("解码CA证书失败")
	}

	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析CA证书失败: %w", err)
	}

	cm.caKey = caKey
	cm.caCert = caCert

	logger.Info("CA证书加载成功")
	return nil
}

// GetCertificate 获取指定域名的证书
// GetCertificate gets certificate for specified domain
func (cm *CertManager) GetCertificate(domain string) (*tls.Certificate, error) {
	// 检查缓存
	if cm.enableCache {
		cm.cacheMutex.RLock()
		if cert, exists := cm.certCache[domain]; exists {
			cm.cacheMutex.RUnlock()
			return cert, nil
		}
		cm.cacheMutex.RUnlock()
	}

	// 生成新证书
	cert, err := cm.generateServerCert(domain)
	if err != nil {
		return nil, fmt.Errorf("生成服务器证书失败: %w", err)
	}

	// 添加到缓存
	if cm.enableCache {
		cm.cacheMutex.Lock()
		cm.certCache[domain] = cert
		cm.cacheMutex.Unlock()

		// 设置过期清理
		go func() {
			time.Sleep(cm.cacheTTL)
			cm.cacheMutex.Lock()
			delete(cm.certCache, domain)
			cm.cacheMutex.Unlock()
		}()
	}

	return cert, nil
}

// generateServerCert 生成服务器证书
// generateServerCert generates server certificate
func (cm *CertManager) generateServerCert(domain string) (*tls.Certificate, error) {
	// 生成服务器私钥
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成服务器私钥失败: %w", err)
	}

	// 创建服务器证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"HackMITM"},
			OrganizationalUnit: []string{"HackMITM Server"},
			CommonName:         domain,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{domain},
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
	}

	// 如果域名是IP地址，添加到IPAddresses字段
	if ip := net.ParseIP(domain); ip != nil {
		template.IPAddresses = []net.IP{ip}
	}

	// 使用CA证书签名服务器证书
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &template, cm.caCert, &serverKey.PublicKey, cm.caKey)
	if err != nil {
		return nil, fmt.Errorf("创建服务器证书失败: %w", err)
	}

	// 构建TLS证书
	cert := &tls.Certificate{
		Certificate: [][]byte{serverCertBytes},
		PrivateKey:  serverKey,
	}

	logger.Debugf("为域名 %s 生成服务器证书成功", domain)
	return cert, nil
}

// GetCACert 获取CA证书内容
// GetCACert returns CA certificate content
func (cm *CertManager) GetCACert() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cm.caCert.Raw,
	})
}

// ExportCACert 导出CA证书到文件
// ExportCACert exports CA certificate to file
func (cm *CertManager) ExportCACert(outputPath string) error {
	caCertPEM := cm.GetCACert()
	if err := os.WriteFile(outputPath, caCertPEM, 0644); err != nil {
		return fmt.Errorf("导出CA证书失败: %w", err)
	}

	logger.Infof("CA证书已导出到: %s", outputPath)
	return nil
}

// ClearCache 清空证书缓存
// ClearCache clears certificate cache
func (cm *CertManager) ClearCache() {
	cm.cacheMutex.Lock()
	defer cm.cacheMutex.Unlock()

	for k := range cm.certCache {
		delete(cm.certCache, k)
	}

	logger.Info("证书缓存已清空")
}

// GetCacheStats 获取缓存统计信息
// GetCacheStats returns cache statistics
func (cm *CertManager) GetCacheStats() map[string]interface{} {
	cm.cacheMutex.RLock()
	defer cm.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_enabled": cm.enableCache,
		"cache_size":    len(cm.certCache),
		"cache_ttl":     cm.cacheTTL.String(),
	}
}
