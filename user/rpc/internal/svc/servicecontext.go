package svc

import (
	"log"
	"os"
	"time"

	"github.com/hd2yao/ecshop/common/mail"
	"github.com/hd2yao/ecshop/common/oss"
	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/user/rpc/internal/config"
)

type ServiceContext struct {
	Config      config.Config
	MailService *mail.MailService
	OssClient   *oss.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化Redis连接池
	if err := redisPool.InitRedisPoolFromGoZero(c.DataRedis); err != nil {
		log.Fatalf("初始化Redis连接池失败: %v", err)
	}

	// 初始化邮件服务 - 从环境变量读取配置
	mailConfig := c.Mail
	if username := os.Getenv("MAIL_USERNAME"); username != "" {
		mailConfig.SMTP.Username = username
	}
	if password := os.Getenv("MAIL_PASSWORD"); password != "" {
		mailConfig.SMTP.Password = password
	}
	if from := os.Getenv("MAIL_FROM"); from != "" {
		mailConfig.SMTP.From = from
	}

	mailService, err := mail.NewMailService(mailConfig, "user", 10*time.Minute)
	if err != nil {
		log.Fatalf("初始化邮件服务失败: %v", err)
	}

	// 初始化OSS客户端
	ossClient, err := oss.NewClient(&c.Oss)
	if err != nil {
		log.Fatalf("初始化OSS客户端失败: %v", err)
	}

	return &ServiceContext{
		Config:      c,
		MailService: mailService,
		OssClient:   ossClient,
	}
}
