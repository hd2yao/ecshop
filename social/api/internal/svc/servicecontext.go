package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/social/api/internal/config"
	"github.com/hd2yao/ecshop/social/rpc/socialclient"
)

type ServiceContext struct {
	Config    config.Config
	SocialRpc socialclient.Social
	JWTAuth   rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		SocialRpc: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		JWTAuth:   middleware.JWTAuthMiddleware(),
	}
}
