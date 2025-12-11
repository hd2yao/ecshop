package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserRelationModel = (*customUserRelationModel)(nil)

type (
	// UserRelationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserRelationModel.
	UserRelationModel interface {
		userRelationModel
	}

	customUserRelationModel struct {
		*defaultUserRelationModel
	}
)

// NewUserRelationModel returns a model for the database table.
func NewUserRelationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserRelationModel {
	return &customUserRelationModel{
		defaultUserRelationModel: newUserRelationModel(conn, c, opts...),
	}
}
