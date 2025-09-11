package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterCaptchaRoutes 注册验证码相关路由
func RegisterCaptchaRoutes(router *gin.Engine) {
	captchaHandler := NewCaptchaHandler()
	
	// 验证码相关API分组
	captchaGroup := router.Group("/api/captcha")
	{
		// 生成验证码
		captchaGroup.POST("/generate", captchaHandler.GenerateCaptcha)
		
		// 验证验证码
		captchaGroup.POST("/verify", captchaHandler.VerifyCaptcha)
		
		// 获取支持的验证码类型
		captchaGroup.GET("/types", captchaHandler.GetCaptchaTypes)
		
		// 获取预设配置
		captchaGroup.GET("/presets", captchaHandler.GetPresetConfigs)
	}
}

// SetupRoutes 设置所有路由
func SetupRoutes() *gin.Engine {
	router := gin.Default()
	
	// 注册验证码路由
	RegisterCaptchaRoutes(router)
	
	// 这里可以注册其他模块的路由
	// RegisterUserRoutes(router)
	// RegisterOrderRoutes(router)
	
	return router
}
