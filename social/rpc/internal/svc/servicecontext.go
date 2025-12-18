package svc

import (
	"log"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/social/model"
	"github.com/hd2yao/ecshop/social/rpc/internal/config"
	"github.com/hd2yao/ecshop/user/rpc/userclient"
)

type ServiceContext struct {
	Config             config.Config
	UserAttentionModel model.UserAttentionModel
	UserFollowerModel  model.UserFollowerModel
	UserRelationModel  model.UserRelationModel
	UserRpc            userclient.User
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

	// 初始化模型（同时用于数据缓存和业务数据存储）
	userAttentionModel := model.NewUserAttentionModel(conn, c.CacheRedis)
	userFollowerModel := model.NewUserFollowerModel(conn, c.CacheRedis)
	userRelationModel := model.NewUserRelationModel(conn, c.CacheRedis)

	// 初始化 User RPC 客户端
	userRpc := userclient.NewUser(zrpc.MustNewClient(c.UserRpc))

	return &ServiceContext{
		Config:             c,
		UserAttentionModel: userAttentionModel,
		UserFollowerModel:  userFollowerModel,
		UserRelationModel:  userRelationModel,
		UserRpc:            userRpc,
	}
}
