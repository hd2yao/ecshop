package oss

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

// Client OSS 客户端封装
type Client struct {
	config    *Config
	ossClient *oss.Client
}

// NewClient 创建 OSS 客户端
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	ossClient, err := oss.New(config.Endpoint, config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create oss client: %v", err)
	}

	return &Client{
		config:    config,
		ossClient: ossClient,
	}, nil
}

// Upload 通用上传方法
func (c *Client) Upload(file io.Reader, filename string, options *UploadOptions) (*UploadResult, error) {
	if options == nil {
		options = &UploadOptions{}
	}

	// 设置默认值
	if options.Dir == "" {
		options.Dir = c.config.UploadPath
	}

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 验证文件
	if err := c.validateFile(content, filename, options); err != nil {
		return nil, err
	}

	// 生成存储路径
	key := c.generateKey(filename, options)

	// 获取存储桶
	bucket, err := c.GetBucket()
	if err != nil {
		return nil, err
	}

	// 设置内容类型
	contentType := c.getContentType(filename)

	// 执行上传
	err = bucket.PutObject(key, bytes.NewReader(content), oss.ContentType(contentType))
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}

	// 构建返回结果
	result := &UploadResult{
		Key:      key,
		Size:     int64(len(content)),
		MimeType: contentType,
		Url:      c.config.GetFileURL(key),
		CDNUrl:   c.config.GetCDNURL(key),
	}

	return result, nil
}

// GetBucket 获取存储桶
func (c *Client) GetBucket() (*oss.Bucket, error) {
	return c.ossClient.Bucket(c.config.DefaultBucket)
}

// GetFileURL 获取文件访问 URL
func (c *Client) GetFileURL(key string) string {
	return c.config.GetFileURL(key)
}

// GetCDNURL 获取 CDN URL
func (c *Client) GetCDNURL(key string) string {
	return c.config.GetCDNURL(key)
}

// CheckFileExists 检查文件是否存在
func (c *Client) CheckFileExists(key string) (bool, error) {
	bucket, err := c.GetBucket()
	if err != nil {
		return false, err
	}

	exists, err := bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %v", err)
	}

	return exists, nil
}

// Copy 复制文件
func (c *Client) Copy(sourceKey, destKey string) (string, error) {
	bucket, err := c.GetBucket()
	if err != nil {
		return "", fmt.Errorf("failed to get bucket: %v", err)
	}

	// 执行复制操作
	_, err = bucket.CopyObject(sourceKey, destKey)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %v", err)
	}

	// 返回新文件的URL
	finalUrl := c.GetFileURL(destKey)
	return finalUrl, nil
}

// Delete 删除文件
func (c *Client) Delete(key string) error {
	bucket, err := c.GetBucket()
	if err != nil {
		return fmt.Errorf("failed to get bucket: %v", err)
	}

	err = bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// validateFile 验证文件
func (c *Client) validateFile(content []byte, filename string, options *UploadOptions) error {
	// 检查文件大小
	maxSize := options.MaxSize
	if maxSize == 0 {
		maxSize = c.config.MaxFileSize
	}
	if int64(len(content)) > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", len(content), maxSize)
	}

	// 检查文件类型
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	allowedTypes := options.AllowedTypes
	if len(allowedTypes) == 0 {
		allowedTypes = c.config.AllowedTypes
	}

	if len(allowedTypes) > 0 && !c.isAllowedType(ext, allowedTypes) {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	return nil
}

// generateKey 生成存储路径
func (c *Client) generateKey(filename string, options *UploadOptions) string {
	// 生成唯一文件名
	var finalFilename string
	if options.Filename != "" {
		finalFilename = options.Filename
	} else {
		ext := filepath.Ext(filename)
		finalFilename = fmt.Sprintf("%s%s", uuid.New().String(), ext)
	}

	// 按日期分目录
	datePath := time.Now().Format("2006/01/02")

	// 组合路径
	return fmt.Sprintf("%s/%s/%s", options.Dir, datePath, finalFilename)
}

// getContentType 获取内容类型
func (c *Client) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".mp4":  "video/mp4",
		".zip":  "application/zip",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

// isAllowedType 检查文件类型是否允许
func (c *Client) isAllowedType(fileType string, allowedTypes []string) bool {
	for _, allowedType := range allowedTypes {
		if allowedType == fileType {
			return true
		}
	}
	return false
}
