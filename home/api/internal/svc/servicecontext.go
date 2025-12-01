package svc

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/hd2yao/ecshop/home/api/internal/config"
	"github.com/hd2yao/ecshop/home/rpc/homeclient"
)

type ServiceContext struct {
	Config config.Config

	HomeRpc homeclient.Home
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		HomeRpc: homeclient.NewHome(zrpc.MustNewClient(c.HomeRpc)),
	}
}
