package captcha

import (
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
)

// Captcha 验证码工具
type Captcha struct {
	store  base64Captcha.Store
	driver base64Captcha.Driver
}

// NewCaptcha 创建验证码实例
func NewCaptcha(module string, expiration time.Duration) *Captcha {
	store := NewCaptchaRedisStore(module, expiration)
	driver := NewDriver(GetDefaultCaptchaRequest())

	return &Captcha{
		store:  store,
		driver: driver,
	}
}

// NewCaptchaWithConfig 创建带配置的验证码实例
func NewCaptchaWithConfig(module string, expiration time.Duration, req CaptchaRequest) *Captcha {
	store := NewCaptchaRedisStore(module, expiration)
	driver := NewDriver(req)

	return &Captcha{
		store:  store,
		driver: driver,
	}
}

// Generate 生成验证码
func (c *Captcha) Generate() (id, b64s, answer string, err error) {
	captcha := base64Captcha.NewCaptcha(c.driver, c.store)
	return captcha.Generate()
}

// Verify 验证验证码
func (c *Captcha) Verify(id, answer string, clear bool) bool {
	return c.store.Verify(id, answer, clear)
}

// GetDefaultCaptchaRequest 获取默认验证码请求配置
func GetDefaultCaptchaRequest() CaptchaRequest {
	return CaptchaRequest{
		CaptchaType: "string",
		Config:      GetDefaultConfig(),
		DrawOpts:    GetDefaultDrawOptions(),
	}
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() DriverConfig {
	return DriverConfig{
		Width:   240,
		Height:  80,
		Length:  5,
		BgColor: color.RGBA{255, 255, 255, 255},
	}
}

// GetDefaultDrawOptions 获取默认绘制选项
func GetDefaultDrawOptions() DrawOptions {
	return DrawOptions{
		UseCustomDraw: false,
		DrawText:      true,
		DrawHollow:    false,
		DrawSine:      false,
		DrawSlimLine:  0,
		DrawNoiseText: "",
	}
}
