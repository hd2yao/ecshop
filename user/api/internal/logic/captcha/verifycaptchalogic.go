package captcha

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/captcha"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
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
	if req.CaptchaId == "" || req.Answer == "" {
		l.Error("验证码ID或答案为空")
		return &types.VerifyResponse{
			Code:    400,
			Message: "验证码ID和答案不能为空",
			Valid:   false,
		}, nil
	}

	// 创建验证码实例并验证
	c := captcha.NewCaptcha("user", 5*time.Minute)
	isValid := c.Verify(req.CaptchaId, req.Answer, true)

	message := "验证失败"
	if isValid {
		message = "验证通过"
	}

	l.Infof("验证码验证结果: ID=%s, 结果=%v", req.CaptchaId, isValid)

	return &types.VerifyResponse{
		Code:    200,
		Message: message,
		Valid:   isValid,
	}, nil
}
