package logic

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/captcha"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type SendRegisterMailCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendRegisterMailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendRegisterMailCodeLogic {
	return &SendRegisterMailCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendRegisterMailCodeLogic) SendRegisterMailCode(in *user.SendRegisterMailCodeReq) (*user.SendMailCodeResp, error) {
	// 1. 参数验证
	if in.CaptchaId == "" || in.CaptchaCode == "" {
		return &user.SendMailCodeResp{
			Code:    400,
			Message: "图形验证码ID和验证码不能为空",
		}, nil
	}

	if in.Email == "" {
		return &user.SendMailCodeResp{
			Code:    400,
			Message: "邮箱地址不能为空",
		}, nil
	}

	// 2. 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(in.Email) {
		return &user.SendMailCodeResp{
			Code:    400,
			Message: "邮箱格式不正确",
		}, nil
	}

	// 3. 验证图形验证码
	c := captcha.NewCaptcha("user", 5*time.Minute)
	if !c.Verify(in.CaptchaId, in.CaptchaCode, true) {
		l.Errorf("图形验证码验证失败: captcha_id=%s", in.CaptchaId)
		return &user.SendMailCodeResp{
			Code:    400,
			Message: "图形验证码错误",
		}, nil
	}

	// 4. 检查邮箱是否已注册
	existingUser, err := l.svcCtx.UserModel.FindOneByMail(l.ctx, sql.NullString{String: in.Email, Valid: true})
	if err != nil && err != sql.ErrNoRows {
		l.Errorf("查询用户失败: %v", err)
		return &user.SendMailCodeResp{
			Code:    500,
			Message: "系统错误",
		}, nil
	}
	if existingUser != nil {
		return &user.SendMailCodeResp{
			Code:    400,
			Message: "该邮箱已被注册",
		}, nil
	}

	// 5. 设置默认验证码长度
	codeLength := int(in.CodeLength)
	if codeLength <= 0 {
		codeLength = 6
	}

	// 6. 发送邮件验证码
	code, err := l.svcCtx.MailService.SendVerifyCode(in.Email, codeLength)
	if err != nil {
		l.Errorf("发送邮件验证码失败: %v", err)
		return &user.SendMailCodeResp{
			Code:    500,
			Message: fmt.Sprintf("发送邮件验证码失败: %v", err),
		}, nil
	}

	l.Infof("注册邮件验证码发送成功: email=%s", in.Email)

	return &user.SendMailCodeResp{
		Code:    200,
		Message: "邮件验证码发送成功",
		Email:   in.Email,
		CodeId:  code, // 测试环境返回验证码
	}, nil
}
