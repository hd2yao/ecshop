package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserRelationModel = (*customUserRelationModel)(nil)

type (
	// UserRelationModel 自定义接口
	UserRelationModel interface {
		userRelationModel
		AddFollowCount(ctx context.Context, userId int64, delta int64) error
		AddFollowerCount(ctx context.Context, userId int64, delta int64) error
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

// AddFollowCount 增加关注数（delta 可为正负）
func (m *customUserRelationModel) AddFollowCount(ctx context.Context, userId int64, delta int64) error {
	query := fmt.Sprintf("insert into %s (`user_id`,`attention_count`,`follower_count`) values (?, ?, 0) "+
		"on duplicate key update `attention_count`=GREATEST(0, `attention_count` + VALUES(`attention_count`))", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, userId, delta)
	return err
}

// AddFollowerCount 增加粉丝数（delta 可为正负）
func (m *customUserRelationModel) AddFollowerCount(ctx context.Context, userId int64, delta int64) error {
	query := fmt.Sprintf("insert into %s (`user_id`,`attention_count`,`follower_count`) values (?, 0, ?) "+
		"on duplicate key update `follower_count`=GREATEST(0, `follower_count` + VALUES(`follower_count`))", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, userId, delta)
	return err
}
