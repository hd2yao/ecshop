package upload

import (
	"context"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type UploadPreviewAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadPreviewAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadPreviewAvatarLogic {
	return &UploadPreviewAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadPreviewAvatarLogic) UploadPreviewAvatar(r *http.Request) (resp *types.UploadPreviewAvatarResponse, err error) {
	// 1. 解析multipart/form-data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		l.Errorf("解析表单数据失败: %v", err)
		return nil, errcode.CommonParamError
	}

	// 2. 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		l.Errorf("获取上传文件失败: %v", err)
		return nil, errcode.CommonParamError
	}
	defer file.Close()

	// 3. 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		l.Errorf("读取文件数据失败: %v", err)
		return nil, errcode.CommonServerError
	}

	// 4. 调用RPC服务
	rpcResp, err := l.svcCtx.UserRpc.UploadPreviewAvatar(l.ctx, &user.UploadPreviewAvatarReq{
		FileData: fileData,
		Filename: header.Filename,
		FileSize: int64(len(fileData)),
	})

	if err != nil {
		l.Errorf("调用RPC服务失败: %v", err)
		return nil, errcode.CommonServerError
	}

	// 检查RPC响应的错误码
	if rpcResp.Code != int32(errcode.Success.Code()) {
		return nil, errcode.CommonServerError
	}

	// 5. 转换响应
	return &types.UploadPreviewAvatarResponse{
		Code:       int(rpcResp.Code),
		Message:    rpcResp.Message,
		PreviewKey: rpcResp.PreviewKey,
		PreviewUrl: rpcResp.PreviewUrl,
	}, nil
}
