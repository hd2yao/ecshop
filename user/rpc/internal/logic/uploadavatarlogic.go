package logic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/oss"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type UploadAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UploadAvatar 文件上传
func (l *UploadAvatarLogic) UploadAvatar(in *user.UploadAvatarReq) (*user.UploadAvatarResp, error) {
	// 参数验证
	// if in.UserId <= 0 {
	// 	return &user.UploadAvatarResp{
	// 		Code:    400,
	// 		Message: "用户ID不能为空",
	// 	}, nil
	// }

	if len(in.FileData) == 0 {
		return &user.UploadAvatarResp{
			Code:    400,
			Message: "文件数据不能为空",
		}, nil
	}

	if in.Filename == "" {
		return &user.UploadAvatarResp{
			Code:    400,
			Message: "文件名不能为空",
		}, nil
	}

	// 日志记录
	l.Infof("用户 %d 开始上传头像，文件名: %s，文件大小: %d bytes", in.UserId, in.Filename, in.FileSize)

	// 设置上传选项
	uploadOptions := &oss.UploadOptions{
		Dir: fmt.Sprintf("avatars/user_%d", in.UserId), // 按用户ID分目录
		// 文件名会在OSS工具中自动生成唯一名称
	}

	// 使用bytes.NewReader将[]byte转为io.Reader
	fileReader := bytes.NewReader(in.FileData)

	// 调用OSS工具上传文件
	result, err := l.svcCtx.OssClient.Upload(fileReader, in.Filename, uploadOptions)
	if err != nil {
		l.Errorf("用户 %d 头像上传失败: %v", in.UserId, err)
		return &user.UploadAvatarResp{
			Code:    500,
			Message: fmt.Sprintf("文件上传失败: %v", err),
		}, nil
	}

	// 上传成功，记录日志
	l.Infof("用户 %d 头像上传成功，URL: %s", in.UserId, result.Url)

	// 返回成功结果
	return &user.UploadAvatarResp{
		Code:    200,
		Message: "上传成功",
		Url:     result.Url,
		CdnUrl:  result.CDNUrl,
		Key:     result.Key,
		Size:    result.Size,
	}, nil
}
