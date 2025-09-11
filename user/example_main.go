package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/hd2yao/ecshop/common/utils/captcha"
	"github.com/hd2yao/ecshop/user/api"
)

func main() {
	// 示例1: 基本用法 - 生成简单验证码
	fmt.Println("=== 基本用法示例 ===")
	basicExample()

	// 示例2: 自定义配置 - 生成复杂验证码
	fmt.Println("\n=== 自定义配置示例 ===")
	customExample()

	// 示例3: 启动Web服务
	fmt.Println("\n=== 启动Web服务 ===")
	startWebServer()
}

// 基本用法示例
func basicExample() {
	// 使用默认配置生成字符串验证码
	req := captcha.CaptchaRequest{
		CaptchaType: "string",
		Config:      captcha.GetDefaultConfig(),
		DrawOpts:    captcha.GetDefaultDrawOptions(),
	}

	id, b64s, answer, err := captcha.Generate(req)
	if err != nil {
		log.Printf("生成验证码失败: %v", err)
		return
	}

	fmt.Printf("验证码ID: %s\n", id)
	fmt.Printf("验证码答案: %s\n", answer)
	fmt.Printf("Base64图片数据长度: %d 字符\n", len(b64s))

	// 验证测试
	isValid := captcha.Verify(id, answer)
	fmt.Printf("验证结果: %v\n", isValid)
}

// 自定义配置示例
func customExample() {
	// 生成带干扰线的数学验证码
	req := captcha.CaptchaRequest{
		CaptchaType: "math",
		Config: captcha.DriverConfig{
			Width:   300,
			Height:  100,
			Length:  1,                              // 数学题长度固定为1
			BgColor: color.RGBA{240, 248, 255, 255}, // 淡蓝色背景
		},
		DrawOpts: captcha.DrawOptions{
			UseCustomDraw: true,
			DrawText:      true,
			DrawHollow:    true,
			DrawSine:      true,
			DrawSlimLine:  3,
			DrawNoiseText: "+-*/=",
		},
	}

	id, b64s, answer, err := captcha.Generate(req)
	if err != nil {
		log.Printf("生成数学验证码失败: %v", err)
		return
	}

	fmt.Printf("数学验证码ID: %s\n", id)
	fmt.Printf("数学验证码答案: %s\n", answer)
	fmt.Printf("Base64图片数据长度: %d 字符\n", len(b64s))

	// 生成中文验证码
	chineseReq := captcha.CaptchaRequest{
		CaptchaType: "chinese",
		Config: captcha.DriverConfig{
			Width:   200,
			Height:  80,
			Length:  4,
			Source:  "验证码测试汉字字符池样本",
			BgColor: color.RGBA{255, 255, 255, 255},
		},
		DrawOpts: captcha.DrawOptions{
			UseCustomDraw: true,
			DrawText:      true,
			DrawSine:      true,
		},
	}

	chId, chB64s, chAnswer, err := captcha.Generate(chineseReq)
	if err != nil {
		log.Printf("生成中文验证码失败: %v", err)
		return
	}

	fmt.Printf("中文验证码ID: %s\n", chId)
	fmt.Printf("中文验证码答案: %s\n", chAnswer)
	fmt.Printf("Base64图片数据长度: %d 字符\n", len(chB64s))
}

// 启动Web服务示例
func startWebServer() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	router := api.SetupRoutes()

	// 添加健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "captcha-service",
			"version": "1.0.0",
		})
	})

	// 添加首页说明
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "验证码服务API",
			"endpoints": map[string]string{
				"POST /api/captcha/generate": "生成验证码",
				"POST /api/captcha/verify":   "验证验证码",
				"GET  /api/captcha/types":    "获取验证码类型",
				"GET  /api/captcha/presets":  "获取预设配置",
				"GET  /health":               "健康检查",
			},
			"examples": map[string]interface{}{
				"generate_simple": map[string]interface{}{
					"captcha_type": "string",
					"config": map[string]interface{}{
						"width":  240,
						"height": 80,
						"length": 5,
					},
					"draw_options": map[string]interface{}{
						"use_custom_draw": false,
						"draw_text":       true,
					},
				},
				"generate_complex": map[string]interface{}{
					"captcha_type": "math",
					"config": map[string]interface{}{
						"width":  300,
						"height": 100,
					},
					"draw_options": map[string]interface{}{
						"use_custom_draw": true,
						"draw_text":       true,
						"draw_hollow":     true,
						"draw_sine":       true,
						"draw_slim_line":  2,
						"draw_noise_text": "123",
					},
				},
			},
		})
	})

	// 启动服务器
	port := ":8080"
	fmt.Printf("验证码服务启动成功，监听端口: %s\n", port)
	fmt.Printf("API文档地址: http://localhost%s/\n", port)
	fmt.Printf("健康检查: http://localhost%s/health\n", port)
	fmt.Printf("生成验证码: POST http://localhost%s/api/captcha/generate\n", port)
	fmt.Printf("验证验证码: POST http://localhost%s/api/captcha/verify\n", port)

	if err := router.Run(port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
