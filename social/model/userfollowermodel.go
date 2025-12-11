package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserFollowerModel = (*customUserFollowerModel)(nil)

type (
	// UserFollowerModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserFollowerModel.
	UserFollowerModel interface {
		userFollowerModel
	}

	customUserFollowerModel struct {
		*defaultUserFollowerModel
	}
)

// NewUserFollowerModel returns a model for the database table.
func NewUserFollowerModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserFollowerModel {
	return &customUserFollowerModel{
		defaultUserFollowerModel: newUserFollowerModel(conn, c, opts...),
	}
}
