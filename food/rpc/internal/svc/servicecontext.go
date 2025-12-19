package svc

import (
	"log"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/food/model"
	"github.com/hd2yao/ecshop/food/rpc/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	FoodModel model.FoodModel
	FoodCache *redisPool.RedisCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Redis 连接池（从 CacheRedis 配置中提取第一个 Redis 配置）
	if len(c.CacheRedis) > 0 {
		redisConf := c.CacheRedis[0]
		if err := redisPool.InitRedisPoolFromGoZero(redisConf.RedisConf); err != nil {
			log.Fatalf("初始化Redis连接池失败: %v", err)
		}
	}

	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DataSource)

	// 初始化 Redis 缓存
	foodCache := redisPool.NewRedisCache("food", "info")

	// 初始化美食模型（同时用于数据缓存和业务数据存储）
	foodModel := model.NewFoodModel(conn, c.CacheRedis, foodCache)

	return &ServiceContext{
		Config:    c,
		FoodModel: foodModel,
		FoodCache: foodCache,
	}
}
