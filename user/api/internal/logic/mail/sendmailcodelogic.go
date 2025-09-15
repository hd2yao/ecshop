package mail

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type SendMailCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMailCodeLogic {
	return &SendMailCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMailCodeLogic) SendMailCode(req *types.SendMailCodeRequest) (resp *types.SendMailCodeResponse, err error) {
	// 调用user-rpc服务发送邮件验证码
	rpcReq := &user.SendMailCodeReq{
		Email:      req.Email,
		CodeLength: int32(req.CodeLength),
	}

	rpcResp, err := l.svcCtx.UserRpc.SendMailCode(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC发送邮件验证码失败: %v", err)
		return &types.SendMailCodeResponse{
			Code:    500,
			Message: "发送邮件验证码失败",
		}, nil
	}

	return &types.SendMailCodeResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		Email:   rpcResp.Email,
		CodeId:  rpcResp.CodeId,
	}, nil
}
