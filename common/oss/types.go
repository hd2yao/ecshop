package oss

// UploadResult 上传结果
type UploadResult struct {
	Url      string `json:"url"`       // 文件访问 URL
	CDNUrl   string `json:"cdn_url"`   // CDN 访问 URL
	Key      string `json:"key"`       // OSS 存储路径
	Size     int64  `json:"size"`      // 文件大小
	MimeType string `json:"mime_type"` // 文件类型
}

// UploadOptions 上传配置选项
type UploadOptions struct {
	Dir          string   `json:"dir"`           // 存储目录
	Filename     string   `json:"filename"`      // 文件名，为空时自动生成
	MaxSize      int64    `json:"max_size"`      // 最大文件大小(字节)
	AllowedTypes []string `json:"allowed_types"` // 允许的文件类型
}