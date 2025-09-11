package captcha

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

// CustomOptions 自定义绘制选项
type CustomOptions struct {
	DrawText      bool
	DrawHollow    bool
	DrawSine      bool
	DrawSlimLine  int
	DrawNoiseText string
}

// DriverCustom 自定义 Driver
type DriverCustom struct {
	Driver  base64Captcha.Driver // 基础 Driver (string/digit/chinese/math)
	Width   int
	Height  int
	BgColor color.RGBA
	Options CustomOptions
}

// NewCustomDriver 构造方法
func NewCustomDriver(width, height int, bgColor color.RGBA, baseDriver base64Captcha.Driver) *DriverCustom {
	return &DriverCustom{
		Driver:  baseDriver,
		Width:   width,
		Height:  height,
		BgColor: bgColor,
		Options: CustomOptions{
			DrawText: true,
		},
	}
}

// 链式配置

func (d *DriverCustom) WithText(enable bool) *DriverCustom {
	d.Options.DrawText = enable
	return d
}

func (d *DriverCustom) WithHollowLine(enable bool) *DriverCustom {
	d.Options.DrawHollow = enable
	return d
}

func (d *DriverCustom) WithSineLine(enable bool) *DriverCustom {
	d.Options.DrawSine = enable
	return d
}

func (d *DriverCustom) WithSlimLine(num int) *DriverCustom {
	d.Options.DrawSlimLine = num
	return d
}

func (d *DriverCustom) WithNoise(text string) *DriverCustom {
	d.Options.DrawNoiseText = text
	return d
}

// GenerateIdQuestionAnswer 调用基础 Driver
func (d *DriverCustom) GenerateIdQuestionAnswer() (id, q, a string) {
	return d.Driver.GenerateIdQuestionAnswer()
}

// DrawCaptcha 使用 ItemChar 绘制
func (d *DriverCustom) DrawCaptcha(content string) (base64Captcha.Item, error) {
	img := NewItemChar(d.Width, d.Height, d.BgColor)
	if d.Options.DrawText && len(Fonts) > 0 {
		_ = img.drawText(content, Fonts)
	}
	if d.Options.DrawHollow {
		img.drawHollowLine()
	}
	if d.Options.DrawSine {
		img.drawSineLine()
	}
	if d.Options.DrawSlimLine > 0 {
		img.drawSlimLine(d.Options.DrawSlimLine)
	}
	if d.Options.DrawNoiseText != "" && len(Fonts) > 0 {
		_ = img.drawNoise(d.Options.DrawNoiseText, Fonts)
	}
	return img, nil
}
