package config

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

// SwaggerConfig Swagger 配置结构体
type SwaggerConfig struct {
	Host     string
	BasePath string
	Schemes  []string
}

// InitSwagger 初始化Swagger
func InitSwagger(r *gin.Engine, config SwaggerConfig) {
	// 设置Swagger信息
	swag.Register(swag.Name, &swag.Spec{
		Host:     config.Host,
		BasePath: config.BasePath,
		Schemes:  config.Schemes,
	})

	// 注册Swagger路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// DefaultConfig 默认Swagger配置
func DefaultConfig() SwaggerConfig {
	return SwaggerConfig{
		Host:     "localhost:8080",
		BasePath: "/api/v1",
		Schemes:  []string{"http", "https"},
	}
}

// WithHost 设置主机
func (c SwaggerConfig) WithHost(host string) SwaggerConfig {
	c.Host = host
	return c
}

// WithBasePath 设置基础路径
func (c SwaggerConfig) WithBasePath(basePath string) SwaggerConfig {
	c.BasePath = basePath
	return c
}
