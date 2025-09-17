package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/hd2yao/ecshop/common/mail"
	"github.com/hd2yao/ecshop/common/oss"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string          `json:"dataSource"` // 数据库连接字符串
	CacheRedis cache.CacheConf `json:"cacheRedis"` // 统一的Redis缓存配置

	Captcha struct {
		StoreType  string `json:",default=redis"`
		Expiration int    `json:",default=600"`
		Prefix     string `json:",default=captcha:"`
	}
	Mail mail.MailConfig
	Oss  oss.Config
}
