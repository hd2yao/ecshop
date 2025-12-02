package captcha

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type VerifyCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewVerifyCaptchaLogic 验证验证码
func NewVerifyCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyCaptchaLogic {
	return &VerifyCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyCaptchaLogic) VerifyCaptcha(req *types.VerifyRequest) (resp *types.VerifyResponse, err error) {
	// 调用user-rpc服务验证验证码
	rpcReq := &user.VerifyCaptchaReq{
		CaptchaId: req.CaptchaId,
		Answer:    req.Answer,
	}

	rpcResp, err := l.svcCtx.UserRpc.VerifyCaptcha(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC验证验证码失败: %v", err)
		return nil, errcode.CommonServerError
	}

	// 检查RPC响应的错误码
	if rpcResp.Code != int32(errcode.Success.Code()) && !rpcResp.Valid {
		// 验证失败，返回验证码错误
		return nil, errcode.UserCodeCaptchaError
	}

	return &types.VerifyResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		Valid:   rpcResp.Valid,
	}, nil
}
