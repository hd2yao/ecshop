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
		// CountAttention 返回用户关注数（user_attention 表计数）
		CountAttention(ctx context.Context, userId int64) (int64, error)
		// CountFollower 返回用户粉丝数（user_follower 表计数）
		CountFollower(ctx context.Context, userId int64) (int64, error)
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

// CountAttention 返回用户关注数（直接从 user_attention 表统计，不使用缓存）
func (m *customUserRelationModel) CountAttention(ctx context.Context, userId int64) (int64, error) {
	countQuery := "select count(1) from user_attention where `user_id`=? and `is_del`=0"
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, userId); err != nil {
		return 0, err
	}
	return total, nil
}

// CountFollower 返回用户粉丝数（直接从 user_follower 表统计，不使用缓存）
func (m *customUserRelationModel) CountFollower(ctx context.Context, userId int64) (int64, error) {
	countQuery := "select count(1) from user_follower where `user_id`=? and `is_del`=0"
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, userId); err != nil {
		return 0, err
	}
	return total, nil
}
