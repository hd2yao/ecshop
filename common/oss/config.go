package oss

import "fmt"

// Config OSS配置
type Config struct {
	// 基础配置
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Endpoint        string `json:"endpoint"`       // OSS服务地址
	DefaultBucket   string `json:"default_bucket"` // 默认存储桶

	// CDN配置
	CDNDomain   string `json:"cdn_domain"`   // CDN域名
	EnableCDN   bool   `json:"enable_cdn"`   // 是否启用CDN
	EnableHTTPS bool   `json:"enable_https"` // 是否启用HTTPS

	// 上传配置
	MaxFileSize  int64    `json:"max_file_size"` // 最大文件大小(字节)
	AllowedTypes []string `json:"allowed_types"` // 全局允许的文件类型
	UploadPath   string   `json:"upload_path"`   // 上传路径前缀
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.AccessKeyId == "" {
		return fmt.Errorf("access_key_id is required")
	}
	if c.AccessKeySecret == "" {
		return fmt.Errorf("access_key_secret is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.DefaultBucket == "" {
		return fmt.Errorf("default_bucket is required")
	}
	if c.MaxFileSize <= 0 {
		c.MaxFileSize = 100 * 1024 * 1024 // 默认100MB
	}
	if c.UploadPath == "" {
		c.UploadPath = "uploads"
	}
	return nil
}

// GetFileURL 获取文件访问URL
func (c *Config) GetFileURL(key string) string {
	protocol := "http"
	if c.EnableHTTPS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s.%s/%s", protocol, c.DefaultBucket, c.Endpoint, key)
}

// GetCDNURL 获取 CDN URL
func (c *Config) GetCDNURL(key string) string {
	if !c.EnableCDN || c.CDNDomain == "" {
		return ""
	}

	protocol := "http"
	if c.EnableHTTPS {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s/%s", protocol, c.CDNDomain, key)
}
