package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAttentionModel = (*customUserAttentionModel)(nil)

type (
	// UserAttentionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAttentionModel.
	UserAttentionModel interface {
		userAttentionModel
	}

	customUserAttentionModel struct {
		*defaultUserAttentionModel
	}
)

// NewUserAttentionModel returns a model for the database table.
func NewUserAttentionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserAttentionModel {
	return &customUserAttentionModel{
		defaultUserAttentionModel: newUserAttentionModel(conn, c, opts...),
	}
}
