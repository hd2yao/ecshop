package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/food/api/internal/config"
	"github.com/hd2yao/ecshop/food/rpc/foodclient"
)

type ServiceContext struct {
	Config  config.Config
	FoodRpc foodclient.Food
	// JWT认证中间件
	JWTAuth rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		FoodRpc: foodclient.NewFood(zrpc.MustNewClient(c.FoodRpc)),
		// 注册JWT认证中间件
		JWTAuth: middleware.JWTAuthMiddleware(),
	}
}
