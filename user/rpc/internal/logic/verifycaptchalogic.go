package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/captcha"
	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type VerifyCaptchaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyCaptchaLogic {
	return &VerifyCaptchaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VerifyCaptcha 验证图形验证码
func (l *VerifyCaptchaLogic) VerifyCaptcha(in *user.VerifyCaptchaReq) (*user.VerifyCaptchaResp, error) {
	// 参数验证
	if in.CaptchaId == "" || in.Answer == "" {
		l.Error("验证码ID或答案为空")
		return &user.VerifyCaptchaResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: errcode.CommonParamError.Msg(),
			Valid:   false,
		}, nil
	}

	// 创建验证码实例并验证
	c := captcha.NewCaptcha("user", 5*time.Minute)
	isValid := c.Verify(in.CaptchaId, in.Answer, true)

	var code int32
	var message string
	if isValid {
		code = int32(errcode.Success.Code())
		message = "验证通过"
	} else {
		code = int32(errcode.UserCodeCaptchaError.Code())
		message = errcode.UserCodeCaptchaError.Msg()
	}

	l.Infof("验证码验证结果: ID=%s, 结果=%v", in.CaptchaId, isValid)

	return &user.VerifyCaptchaResp{
		Code:    code,
		Message: message,
		Valid:   isValid,
	}, nil
}
