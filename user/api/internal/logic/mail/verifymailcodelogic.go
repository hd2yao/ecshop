package mail

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type VerifyMailCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyMailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyMailCodeLogic {
	return &VerifyMailCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyMailCodeLogic) VerifyMailCode(req *types.VerifyMailCodeRequest) (resp *types.VerifyMailCodeResponse, err error) {
	// 调用user-rpc服务验证邮件验证码
	rpcReq := &user.VerifyMailCodeReq{
		Email: req.Email,
		Code:  req.Code,
	}

	rpcResp, err := l.svcCtx.UserRpc.VerifyMailCode(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC验证邮件验证码失败: %v", err)
		return &types.VerifyMailCodeResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
			Valid:   false,
		}, nil
	}

	return &types.VerifyMailCodeResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		Valid:   rpcResp.Valid,
	}, nil
}
