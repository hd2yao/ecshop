package logic

import (
	"context"

	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMailCodeLogic {
	return &SendMailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendMailCode 发送邮件验证码
func (l *SendMailCodeLogic) SendMailCode(in *user.SendMailCodeReq) (*user.SendMailCodeResp, error) {
	// 参数验证
	if in.Email == "" {
		return &user.SendMailCodeResp{
			Code:    400,
			Message: "邮箱地址不能为空",
		}, nil
	}

	// 设置默认验证码长度
	codeLength := int(in.CodeLength)
	if codeLength <= 0 {
		codeLength = 6
	}

	// 发送邮件验证码
	code, err := l.svcCtx.MailService.SendVerifyCode(in.Email, codeLength)
	if err != nil {
		l.Errorf("发送邮件验证码失败: %v", err)
		return &user.SendMailCodeResp{
			Code:    500,
			Message: "发送邮件验证码失败: " + err.Error(),
		}, nil
	}

	l.Infof("邮件验证码发送成功: email=%s", in.Email)

	return &user.SendMailCodeResp{
		Code:    200,
		Message: "邮件验证码发送成功",
		Email:   in.Email,
		CodeId:  code, // 测试环境返回验证码
	}, nil
}
