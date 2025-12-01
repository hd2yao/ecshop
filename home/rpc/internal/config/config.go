package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// 数据库连接（目前可选，预留扩展）
	DataSource string          `json:"dataSource"`
	// 统一的 Redis 缓存配置（与 user 服务保持一致）
	CacheRedis cache.CacheConf `json:"cacheRedis"`
}
