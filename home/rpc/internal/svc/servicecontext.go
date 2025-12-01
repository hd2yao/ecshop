package svc

import (
	"log"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/home/rpc/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	FeedCache *redisPool.RedisCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Redis 连接池（从 CacheRedis 配置中提取第一个节点）
	if len(c.CacheRedis) > 0 {
		redisConf := c.CacheRedis[0]
		if err := redisPool.InitRedisPoolFromGoZero(redisConf.RedisConf); err != nil {
			log.Fatalf("初始化 Redis 连接池失败: %v", err)
		}
	}

	return &ServiceContext{
		Config:    c,
		FeedCache: redisPool.NewRedisCache("home", "feed"),
	}
}
