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
		return nil, errcode.CommonServerError
	}

	// 检查RPC响应的错误码
	if rpcResp.Code != int32(errcode.Success.Code()) {
		// 根据RPC返回的错误码返回对应的AppError
		if appErr := l.getAppErrorByCode(int(rpcResp.Code)); appErr != nil {
			return nil, appErr
		}
		return nil, errcode.CommonServerError
	}

	return &types.SendMailCodeResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		Email:   rpcResp.Email,
		CodeId:  rpcResp.CodeId,
	}, nil
}

// getAppErrorByCode 根据错误码获取对应的AppError
func (l *SendRegisterMailCodeLogic) getAppErrorByCode(code int) *errcode.AppError {
	switch code {
	case errcode.CommonParamError.Code():
		return errcode.CommonParamError
	case errcode.CommonServerError.Code():
		return errcode.CommonServerError
	case errcode.UserCodeCaptchaError.Code():
		return errcode.UserCodeCaptchaError
	case errcode.UserAccountExist.Code():
		return errcode.UserAccountExist
	default:
		return nil
	}
}
