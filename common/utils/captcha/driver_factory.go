package captcha

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

// NewDriver 根据字符串返回对应 driver
func NewDriver(driverType string, opt CustomOptions) base64Captcha.Driver {
	switch driverType {
	case "audio":
		// 音频验证码：不走图片渲染，直接使用官方 audio driver
		return base64Captcha.NewDriverAudio(6, "en")
	case "digit":
		// 内容由官方 digit 生成，绘制统一走我们自定义 ItemChar
		base := base64Captcha.NewDriverDigit(
			80,  // 高
			240, // 宽
			5,   // 长度（用于生成内容）
			0.7, // 噪点强度（仅影响官方绘制，但我们不会用）
			80,  // 背景噪点数量（同上）
		)
		return NewCustomDriver(240, 80, color.RGBA{255, 255, 255, 255}, base).
			WithText(opt.DrawText).
			WithHollowLine(opt.DrawHollow).
			WithSineLine(opt.DrawSine).
			WithSlimLine(opt.DrawSlimLine).
			WithNoise(opt.DrawNoiseText)

	case "math":
		// 数学题内容生成交给官方 math driver；渲染由我们来画
		base := base64Captcha.NewDriverMath(
			80, 240, 0,
			// showLineOptions / BgColor / FontsStorage / fonts 在我们自定义渲染中不使用
			base64Captcha.OptionShowSlimeLine,
			&color.RGBA{255, 255, 255, 255},
			base64Captcha.DefaultEmbeddedFonts,
			[]string{},
		)
		return NewCustomDriver(240, 80, color.RGBA{255, 255, 255, 255}, base).
			WithText(opt.DrawText).
			WithHollowLine(opt.DrawHollow).
			WithSineLine(opt.DrawSine).
			WithSlimLine(opt.DrawSlimLine).
			WithNoise(opt.DrawNoiseText)
	case "chinese":
		// 中文内容生成交给官方 chinese driver；渲染由我们来画
		base := base64Captcha.NewDriverChinese(
			80, 240, 5, base64Captcha.OptionShowSlimeLine, 4,
			"验证码测试中文随便取一些字做池",
			&color.RGBA{240, 240, 240, 255},
			base64Captcha.DefaultEmbeddedFonts,
			[]string{},
		)
		return NewCustomDriver(240, 80, color.RGBA{255, 255, 255, 255}, base).
			WithText(opt.DrawText).
			WithHollowLine(opt.DrawHollow).
			WithSineLine(opt.DrawSine).
			WithSlimLine(opt.DrawSlimLine).
			WithNoise(opt.DrawNoiseText)
	case "string":
		fallthrough
	default: // 默认 string
		// 字符串内容生成交给官方 string driver；渲染统一走我们自定义
		base := base64Captcha.NewDriverString(
			80, 240, 0,
			base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowHollowLine,
			5,
			"1234567890ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz中文验证码",
			nil,
			base64Captcha.DefaultEmbeddedFonts,
			[]string{},
		)
		return NewCustomDriver(240, 80, color.RGBA{255, 255, 255, 255}, base).
			WithText(opt.DrawText).
			WithHollowLine(opt.DrawHollow).
			WithSineLine(opt.DrawSine).
			WithSlimLine(opt.DrawSlimLine).
			WithNoise(opt.DrawNoiseText)
	}
}
