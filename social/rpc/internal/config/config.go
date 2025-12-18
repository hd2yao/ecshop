package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/hd2yao/ecshop/common/rocketmq"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	CacheRedis cache.CacheConf
	RocketMQ   rocketmq.Config `json:",optional"`
	UserRpc    zrpc.RpcClientConf
}
