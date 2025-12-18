package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserFollowerModel = (*customUserFollowerModel)(nil)

type (
	// UserFollowerModel 自定义接口
	UserFollowerModel interface {
		userFollowerModel
		FindOneByUserAndFollower(ctx context.Context, userId, followerId int64) (*UserFollower, error)
		UpsertFollower(ctx context.Context, userId, followerId int64) error
		SoftDeleteFollower(ctx context.Context, userId, followerId int64) error
		ListFollowers(ctx context.Context, userId int64, offset, limit int64) ([]*UserFollower, int64, error)
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

// FindOneByUserAndFollower 根据 user_id + follower_id 查询（无缓存）
func (m *customUserFollowerModel) FindOneByUserAndFollower(ctx context.Context, userId, followerId int64) (*UserFollower, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=? and `follower_id`=? limit 1", userFollowerRows, m.table)
	var resp UserFollower
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, followerId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// UpsertFollower 插入或恢复粉丝关系
func (m *customUserFollowerModel) UpsertFollower(ctx context.Context, userId, followerId int64) error {
	record, err := m.FindOneByUserAndFollower(ctx, userId, followerId)
	if err != nil && err != ErrNotFound {
		return err
	}
	if record == nil || err == ErrNotFound {
		query := fmt.Sprintf("insert into %s (`user_id`,`follower_id`,`is_del`) values (?, ?, 0)", m.table)
		_, err := m.ExecNoCacheCtx(ctx, query, userId, followerId)
		return err
	}
	if record.IsDel != 0 {
		query := fmt.Sprintf("update %s set `is_del`=0 where `id`=?", m.table)
		_, err := m.ExecNoCacheCtx(ctx, query, record.Id)
		return err
	}
	return nil
}

// SoftDeleteFollower 软删除粉丝关系
func (m *customUserFollowerModel) SoftDeleteFollower(ctx context.Context, userId, followerId int64) error {
	query := fmt.Sprintf("update %s set `is_del`=1 where `user_id`=? and `follower_id`=? and `is_del`=0", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, userId, followerId)
	return err
}

// ListFollowers 粉丝列表（按创建时间倒序），返回列表和总数
func (m *customUserFollowerModel) ListFollowers(ctx context.Context, userId int64, offset, limit int64) ([]*UserFollower, int64, error) {
	listQuery := fmt.Sprintf("select %s from %s where `user_id`=? and `is_del`=0 order by `create_time` desc limit ? offset ?", userFollowerRows, m.table)
	var resp []*UserFollower
	if err := m.QueryRowsNoCacheCtx(ctx, &resp, listQuery, userId, limit, offset); err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}
	countQuery := fmt.Sprintf("select count(1) from %s where `user_id`=? and `is_del`=0", m.table)
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, userId); err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}
