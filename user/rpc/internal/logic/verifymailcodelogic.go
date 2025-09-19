package logic

import (
	"context"

	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyMailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyMailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyMailCodeLogic {
	return &VerifyMailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VerifyMailCode 验证邮件验证码
func (l *VerifyMailCodeLogic) VerifyMailCode(in *user.VerifyMailCodeReq) (*user.VerifyMailCodeResp, error) {
	// 参数验证
	if in.Email == "" || in.Code == "" {
		return &user.VerifyMailCodeResp{
			Code:    400,
			Message: "邮箱地址和验证码不能为空",
			Valid:   false,
		}, nil
	}

	// 验证邮件验证码
	isValid := l.svcCtx.MailService.VerifyCode(in.Email, in.Code)

	message := "验证失败"
	if isValid {
		message = "验证通过"
	}

	l.Infof("邮件验证码验证结果: email=%s, valid=%v", in.Email, isValid)

	return &user.VerifyMailCodeResp{
		Code:    200,
		Message: message,
		Valid:   isValid,
	}, nil
}
