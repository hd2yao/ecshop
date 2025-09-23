package mail

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type SendRegisterMailCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendRegisterMailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendRegisterMailCodeLogic {
	return &SendRegisterMailCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SendRegisterMailCode 注册时发送邮件验证码（需要先验证图形验证码）
func (l *SendRegisterMailCodeLogic) SendRegisterMailCode(req *types.SendRegisterMailCodeRequest) (resp *types.SendMailCodeResponse, err error) {
	// 1. 调用RPC服务处理注册邮件验证码发送
	rpcResp, err := l.svcCtx.UserRpc.SendRegisterMailCode(l.ctx, &user.SendRegisterMailCodeReq{
		CaptchaId:   req.CaptchaId,
		CaptchaCode: req.CaptchaCode,
		Email:       req.Email,
		CodeLength:  int32(req.CodeLength),
	})
	if err != nil {
		l.Errorf("RPC调用失败: %v", err)
		return &types.SendMailCodeResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	return &types.SendMailCodeResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		Email:   rpcResp.Email,
		CodeId:  rpcResp.CodeId,
	}, nil
}
