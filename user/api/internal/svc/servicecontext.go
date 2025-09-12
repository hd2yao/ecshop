package svc

import (
	"log"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/user/api/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化Redis连接池
	if err := redisPool.InitRedisPoolFromGoZero(c.Redis); err != nil {
		log.Fatalf("初始化Redis连接池失败: %v", err)
	}

	return &ServiceContext{
		Config: c,
	}
}
