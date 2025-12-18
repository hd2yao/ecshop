package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	SocialRpc zrpc.RpcClientConf
	LogFormat string `json:",optional,default=default"` // 日志格式: default, gin
}
