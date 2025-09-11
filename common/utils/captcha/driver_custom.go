package captcha

import (
	"fmt"
	"image/color"
	"log"

	"github.com/mojocn/base64Captcha"
)

// DriverCustom 自定义 Driver
type DriverCustom struct {
	Driver  base64Captcha.Driver // 基础 Driver (string/digit/chinese/math)
	Width   int
	Height  int
	BgColor color.RGBA
	Options DrawOptions
}

// NewCustomDriver 构造方法
func NewCustomDriver(width, height int, bgColor color.RGBA, baseDriver base64Captcha.Driver, opts DrawOptions) *DriverCustom {
	// 设置默认值
	if !opts.DrawText && !opts.DrawHollow && !opts.DrawSine && opts.DrawSlimLine == 0 && opts.DrawNoiseText == "" {
		opts.DrawText = true // 默认至少绘制文字
	}
	
	return &DriverCustom{
		Driver:  baseDriver,
		Width:   width,
		Height:  height,
		BgColor: bgColor,
		Options: opts,
	}
}


// GenerateIdQuestionAnswer 调用基础 Driver
func (d *DriverCustom) GenerateIdQuestionAnswer() (id, q, a string) {
	return d.Driver.GenerateIdQuestionAnswer()
}

// DrawCaptcha 使用 ItemChar 绘制
func (d *DriverCustom) DrawCaptcha(content string) (base64Captcha.Item, error) {
	img := NewItemChar(d.Width, d.Height, d.BgColor)
	
	// 绘制主要文字内容
	if d.Options.DrawText {
		if HasFonts() {
			if err := img.drawText(content, Fonts); err != nil {
				return nil, fmt.Errorf("绘制文字失败: %v", err)
			}
		} else {
			log.Println("警告: 没有可用字体，跳过文字绘制")
		}
	}
	
	// 绘制各种干扰效果
	if d.Options.DrawHollow {
		img.drawHollowLine()
	}
	if d.Options.DrawSine {
		img.drawSineLine()
	}
	if d.Options.DrawSlimLine > 0 {
		img.drawSlimLine(d.Options.DrawSlimLine)
	}
	
	// 绘制噪点文字
	if d.Options.DrawNoiseText != "" {
		if HasFonts() {
			if err := img.drawNoise(d.Options.DrawNoiseText, Fonts); err != nil {
				log.Printf("绘制噪点失败: %v", err)
			}
		} else {
			log.Println("警告: 没有可用字体，跳过噪点绘制")
		}
	}
	
	return img, nil
}
