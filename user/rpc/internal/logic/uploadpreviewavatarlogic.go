package logic

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/oss"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type UploadPreviewAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadPreviewAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPreviewAvatarLogic {
	return &UploadPreviewAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UploadPreviewAvatar 上传预览头像到临时目录
func (l *UploadPreviewAvatarLogic) UploadPreviewAvatar(in *user.UploadPreviewAvatarReq) (*user.UploadPreviewAvatarResp, error) {
	// 1. 参数验证
	if len(in.FileData) == 0 {
		return &user.UploadPreviewAvatarResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: errcode.CommonParamError.Msg(),
		}, nil
	}

	if in.Filename == "" {
		return &user.UploadPreviewAvatarResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: errcode.CommonParamError.Msg(),
		}, nil
	}

	// 2. 文件格式验证
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	ext := strings.ToLower(filepath.Ext(in.Filename))
	if !allowedExts[ext] {
		return &user.UploadPreviewAvatarResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "不支持的文件格式，只支持 jpg, jpeg, png, gif",
		}, nil
	}

	// 3. 文件大小验证 (限制5MB)
	maxSize := int64(5 * 1024 * 1024)
	if in.FileSize > maxSize {
		return &user.UploadPreviewAvatarResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "文件大小不能超过5MB",
		}, nil
	}

	// 4. 生成临时标识
	tempKey, err := l.generateTempKey()
	if err != nil {
		l.Errorf("生成临时标识失败: %v", err)
		return &user.UploadPreviewAvatarResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 5. 生成临时文件名
	tempFilename := fmt.Sprintf("temp_%s%s", tempKey, ext)

	// 6. 上传到OSS临时目录
	uploadOptions := &oss.UploadOptions{
		Dir: "avatars/temp", // 临时目录
		// 文件名会在OSS工具中自动生成唯一名称
	}

	// 使用bytes.NewReader将[]byte转为io.Reader
	fileReader := bytes.NewReader(in.FileData)

	// 调用OSS工具上传文件
	result, err := l.svcCtx.OssClient.Upload(fileReader, tempFilename, uploadOptions)
	if err != nil {
		l.Errorf("上传到 OSS 失败: %v", err)
		return &user.UploadPreviewAvatarResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "上传失败",
		}, nil
	}

	if err := l.svcCtx.PreviewAvatarCache.Set(l.ctx, tempKey, result.Url, 10*time.Minute); err != nil {
		l.Errorf("保存临时信息失败: %v", err)
		// 不影响上传结果，只记录日志
	}

	// 8. 生成带签名的预览URL（临时文件，30分钟有效期）
	signedURL, err := l.svcCtx.OssClient.GetSignedURL(result.Key, 300)
	previewUrl := signedURL
	if err != nil {
		l.Errorf("生成签名URL失败: %v", err)
		// 如果签名失败，使用CDN或原始地址
		previewUrl = result.CDNUrl
		if previewUrl == "" {
			previewUrl = result.Url
		}
	}

	l.Infof("预览头像上传成功: tempKey=%s, url=%s", tempKey, result.Url)

	return &user.UploadPreviewAvatarResp{
		Code:       int32(errcode.Success.Code()),
		Message:    errcode.Success.Msg(),
		PreviewKey: tempKey,
		PreviewUrl: previewUrl,
	}, nil
}

// generateTempKey 生成临时标识
func (l *UploadPreviewAvatarLogic) generateTempKey() (string, error) {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	randomStr := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%d_%s", timestamp, randomStr), nil
}
