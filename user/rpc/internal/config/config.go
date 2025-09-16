package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/hd2yao/ecshop/common/mail"
	"github.com/hd2yao/ecshop/common/oss"
)

type Config struct {
	zrpc.RpcServerConf
	DataRedis redis.RedisConf `json:"dataRedis"` // 完全独立的Redis配置

	Captcha struct {
		StoreType  string `json:",default=redis"`
		Expiration int    `json:",default=600"`
		Prefix     string `json:",default=captcha:"`
	}
	Mail mail.MailConfig
	Oss  oss.Config
}
