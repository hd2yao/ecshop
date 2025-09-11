package captcha

import (
	"image/color"
	
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// Generate 生成验证码
func Generate(req CaptchaRequest) (id, b64s, answer string, err error) {
	driver := NewDriver(req)
	c := base64Captcha.NewCaptcha(driver, store)
	return c.Generate()
}

// Verify 校验验证码
func Verify(id, value string) bool {
	return store.Verify(id, value, true)
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() DriverConfig {
	return DriverConfig{
		Width:  240,
		Height: 80,
		Length: 5,
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
