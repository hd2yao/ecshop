package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	commonMiddleware "github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/home/api/internal/config"
	"github.com/hd2yao/ecshop/home/rpc/homeclient"
)

type ServiceContext struct {
	Config  config.Config
	JWTAuth rest.Middleware

	HomeRpc homeclient.Home
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		JWTAuth: commonMiddleware.JWTAuthMiddleware(),
		HomeRpc: homeclient.NewHome(zrpc.MustNewClient(c.HomeRpc)),
	}
}
