package svc

import (
	"log"
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/hd2yao/ecshop/common/mail"
	"github.com/hd2yao/ecshop/common/oss"
	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/config"
)

type ServiceContext struct {
	Config      config.Config
	MailService *mail.MailService
	OssClient   *oss.Client
	UserModel   model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DataSource)
	
	// 初始化用户模型（同时用于数据缓存和业务数据存储）
	userModel := model.NewUserModel(conn, c.CacheRedis)

	// 初始化Redis连接池（从CacheRedis配置中提取第一个Redis配置）
	if len(c.CacheRedis) > 0 {
		redisConf := c.CacheRedis[0]
		if err := redisPool.InitRedisPoolFromGoZero(redisConf.RedisConf); err != nil {
			log.Fatalf("初始化Redis连接池失败: %v", err)
		}
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
		UserModel:   userModel,
	}
}
