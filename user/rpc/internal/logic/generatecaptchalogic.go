package logic

import (
	"context"
	"image/color"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/captcha"
	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/rpc/internal/svc"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type GenerateCaptchaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateCaptchaLogic {
	return &GenerateCaptchaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GenerateCaptcha 生成图形验证码
func (l *GenerateCaptchaLogic) GenerateCaptcha(in *user.GenerateCaptchaReq) (*user.GenerateCaptchaResp, error) {
	// 设置默认值
	config := in.Config
	if config == nil {
		config = &user.CaptchaConfig{}
	}
	if config.Width == 0 {
		config.Width = 240
	}
	if config.Height == 0 {
		config.Height = 80
	}
	if config.Length == 0 {
		config.Length = 5
	}
	if config.BgColor == "" {
		config.BgColor = "255,255,255,255"
	}

	captchaType := in.CaptchaType
	if captchaType == "" {
		captchaType = "string"
	}

	// 转换为captcha包的请求格式
	captchaReq := captcha.CaptchaRequest{
		CaptchaType: captchaType,
		Config: captcha.DriverConfig{
			Width:   int(config.Width),
			Height:  int(config.Height),
			Length:  int(config.Length),
			BgColor: color.RGBA{255, 255, 255, 255}, // 默认白色背景
		},
		DrawOpts: captcha.DrawOptions{
			UseCustomDraw: in.DrawOpts.UseCustomDraw,
			DrawText:      in.DrawOpts.DrawText,
			DrawHollow:    in.DrawOpts.DrawHollow,
			DrawSine:      in.DrawOpts.DrawSine,
			DrawSlimLine:  int(in.DrawOpts.DrawSlimLine),
			DrawNoiseText: in.DrawOpts.DrawNoiseText,
		},
	}

	// 生成验证码
	id, b64s, answer, err := generateCaptchaWithConfig("user", 5*time.Minute, captchaReq)
	if err != nil {
		l.Errorf("生成验证码失败: %v", err)
		return &user.GenerateCaptchaResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "验证码生成失败",
		}, nil
	}

	l.Infof("验证码生成成功，ID: %s", id)

	return &user.GenerateCaptchaResp{
		Code:      int32(errcode.Success.Code()),
		Message:   errcode.Success.Msg(),
		CaptchaId: id,
		ImageData: b64s,
		Answer:    answer, // 生产环境不返回答案
	}, nil
}

// generateCaptchaWithConfig 使用配置生成验证码
func generateCaptchaWithConfig(module string, expiration time.Duration, req captcha.CaptchaRequest) (id, b64s, answer string, err error) {
	c := captcha.NewCaptchaWithConfig(module, expiration, req)
	return c.Generate()
}
