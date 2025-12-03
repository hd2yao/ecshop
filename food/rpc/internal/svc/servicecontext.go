package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/hd2yao/ecshop/food/model"
	"github.com/hd2yao/ecshop/food/rpc/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	FoodModel model.FoodModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DataSource)

	// 初始化美食模型
	foodModel := model.NewFoodModel(conn, c.CacheRedis)

	return &ServiceContext{
		Config:    c,
		FoodModel: foodModel,
	}
}
