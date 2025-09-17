package upload

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(req *types.UploadAvatarRequest, file multipart.File, fileHeader *multipart.FileHeader) (resp *types.UploadResponse, err error) {
	// 参数验证
	if req.UserId <= 0 {
		return &types.UploadResponse{
			Code:    400,
			Message: "用户ID不能为空",
		}, nil
	}

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		l.Errorf("读取文件数据失败: %v", err)
		return &types.UploadResponse{
			Code:    500,
			Message: "读取文件数据失败",
		}, nil
	}

	// 构建RPC请求
	rpcReq := &user.UploadAvatarReq{
		UserId:   req.UserId,
		FileData: fileData,
		Filename: fileHeader.Filename,
		FileSize: fileHeader.Size,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.UserRpc.UploadAvatar(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC服务失败: %v", err)
		return &types.UploadResponse{
			Code:    500,
			Message: "服务内部错误",
		}, nil
	}

	// 检查RPC响应
	if rpcResp.Code != 200 {
		return &types.UploadResponse{
			Code:    int(rpcResp.Code),
			Message: rpcResp.Message,
		}, nil
	}

	// 返回成功结果
	return &types.UploadResponse{
		Code:    200,
		Message: "上传成功",
		Url:     rpcResp.Url,
		CDNUrl:  rpcResp.CdnUrl,
		Key:     rpcResp.Key,
		Size:    rpcResp.Size,
	}, nil
}
