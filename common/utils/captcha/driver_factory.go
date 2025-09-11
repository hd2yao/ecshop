package captcha

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

// DriverConfig 驱动器配置
type DriverConfig struct {
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Length int        `json:"length"`
	Source string     `json:"source"`
	BgColor color.RGBA `json:"bg_color"`
}

// DrawOptions 绘制选项
type DrawOptions struct {
	UseCustomDraw bool   `json:"use_custom_draw"` // 是否使用自定义绘制
	DrawText      bool   `json:"draw_text"`
	DrawHollow    bool   `json:"draw_hollow"`
	DrawSine      bool   `json:"draw_sine"`
	DrawSlimLine  int    `json:"draw_slim_line"`
	DrawNoiseText string `json:"draw_noise_text"`
}

// CaptchaRequest 验证码请求参数
type CaptchaRequest struct {
	CaptchaType string       `json:"captcha_type"` // audio, digit, math, chinese, string
	Config      DriverConfig `json:"config"`
	DrawOpts    DrawOptions  `json:"draw_options"`
}

// NewDriver 根据请求参数创建对应的driver
func NewDriver(req CaptchaRequest) base64Captcha.Driver {
	config := req.Config
	// 设置默认值
	if config.Width == 0 {
		config.Width = 240
	}
	if config.Height == 0 {
		config.Height = 80
	}
	if config.Length == 0 {
		config.Length = 5
	}
	// 确保背景色不是透明的
	if config.BgColor.R == 0 && config.BgColor.G == 0 && config.BgColor.B == 0 && config.BgColor.A == 0 {
		config.BgColor = color.RGBA{255, 255, 255, 255} // 默认白色背景
	}

	switch req.CaptchaType {
	case "audio":
		// 音频验证码：直接使用官方实现
		return base64Captcha.NewDriverAudio(config.Length, "en")
		
	case "digit":
		if req.DrawOpts.UseCustomDraw {
			// 使用自定义绘制
			base := base64Captcha.NewDriverDigit(config.Height, config.Width, config.Length, 0.7, 80)
			return NewCustomDriver(config.Width, config.Height, config.BgColor, base, req.DrawOpts)
		}
		// 使用官方绘制
		return base64Captcha.NewDriverDigit(config.Height, config.Width, config.Length, 0.7, 80)
		
	case "math":
		if req.DrawOpts.UseCustomDraw {
			base := base64Captcha.NewDriverMath(config.Height, config.Width, 0,
				base64Captcha.OptionShowSlimeLine, &config.BgColor,
				base64Captcha.DefaultEmbeddedFonts, []string{})
			return NewCustomDriver(config.Width, config.Height, config.BgColor, base, req.DrawOpts)
		}
		return base64Captcha.NewDriverMath(config.Height, config.Width, 0,
			base64Captcha.OptionShowSlimeLine, &config.BgColor,
			base64Captcha.DefaultEmbeddedFonts, []string{})
			
	case "chinese":
		source := config.Source
		if source == "" {
			source = "验证码测试中文随便取一些字做池子用来生成"
		}
		if req.DrawOpts.UseCustomDraw {
			base := base64Captcha.NewDriverChinese(config.Height, config.Width, config.Length,
				base64Captcha.OptionShowSlimeLine, 4, source, &config.BgColor,
				base64Captcha.DefaultEmbeddedFonts, []string{})
			return NewCustomDriver(config.Width, config.Height, config.BgColor, base, req.DrawOpts)
		}
		return base64Captcha.NewDriverChinese(config.Height, config.Width, config.Length,
			base64Captcha.OptionShowSlimeLine, 4, source, &config.BgColor,
			base64Captcha.DefaultEmbeddedFonts, []string{})
			
	case "string":
		fallthrough
	default:
		source := config.Source
		if source == "" {
			source = "1234567890ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz"
		}
		if req.DrawOpts.UseCustomDraw {
			base := base64Captcha.NewDriverString(config.Height, config.Width, 0,
				base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowHollowLine,
				config.Length, source, &config.BgColor,
				base64Captcha.DefaultEmbeddedFonts, []string{})
			return NewCustomDriver(config.Width, config.Height, config.BgColor, base, req.DrawOpts)
		}
		return base64Captcha.NewDriverString(config.Height, config.Width, 0,
			base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowHollowLine,
			config.Length, source, &config.BgColor,
			base64Captcha.DefaultEmbeddedFonts, []string{})
	}
}
