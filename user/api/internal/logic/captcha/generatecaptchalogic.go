package captcha

import (
	"context"
	"image/color"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/captcha"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
)

type GenerateCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGenerateCaptchaLogic 生成验证码
func NewGenerateCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateCaptchaLogic {
	return &GenerateCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateCaptchaLogic) GenerateCaptcha(req *types.CaptchaRequest) (resp *types.CaptchaResponse, err error) {
	// 设置默认值
	if req.Config.Width == 0 {
		req.Config.Width = 240
	}
	if req.Config.Height == 0 {
		req.Config.Height = 80
	}
	if req.Config.Length == 0 {
		req.Config.Length = 5
	}
	if req.Config.BgColor == "" {
		req.Config.BgColor = "255,255,255,255"
	}
	if req.CaptchaType == "" {
		req.CaptchaType = "string"
	}

	// 转换为captcha包的请求格式
	captchaReq := convertToCaptchaRequest(*req)

	// 生成验证码
	id, b64s, answer, err := generateCaptchaWithConfig("user", 5*time.Minute, captchaReq)
	if err != nil {
		l.Errorf("生成验证码失败: %v", err)
		return &types.CaptchaResponse{
			Code:    500,
			Message: "验证码生成失败",
		}, nil
	}

	l.Infof("验证码生成成功，ID: %s", id)

	return &types.CaptchaResponse{
		Code:      200,
		Message:   "验证码生成成功",
		CaptchaId: id,
		ImageData: b64s,
		Answer:    answer, // 生产环境不返回答案
	}, nil
}

// convertToCaptchaRequest 转换API请求到captcha包请求格式
func convertToCaptchaRequest(req types.CaptchaRequest) captcha.CaptchaRequest {
	return captcha.CaptchaRequest{
		CaptchaType: req.CaptchaType,
		Config: captcha.DriverConfig{
			Width:   req.Config.Width,
			Height:  req.Config.Height,
			Length:  req.Config.Length,
			BgColor: color.RGBA{255, 255, 255, 255}, // 默认白色背景
		},
		DrawOpts: captcha.DrawOptions{
			UseCustomDraw: req.DrawOpts.UseCustomDraw,
			DrawText:      req.DrawOpts.DrawText,
			DrawHollow:    req.DrawOpts.DrawHollow,
			DrawSine:      req.DrawOpts.DrawSine,
			DrawSlimLine:  req.DrawOpts.DrawSlimLine,
			DrawNoiseText: req.DrawOpts.DrawNoiseText,
		},
	}
}

// generateCaptchaWithConfig 使用配置生成验证码
func generateCaptchaWithConfig(module string, expiration time.Duration, req captcha.CaptchaRequest) (id, b64s, answer string, err error) {
	c := captcha.NewCaptchaWithConfig(module, expiration, req)
	return c.Generate()
}
