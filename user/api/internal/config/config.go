package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Redis   redis.RedisConf
	Captcha CaptchaConfig
}

type CaptchaConfig struct {
	StoreType  string `json:",default=redis"`
	Expiration int    `json:",default=600"`
	Prefix     string `json:",default=captcha:"`
}
