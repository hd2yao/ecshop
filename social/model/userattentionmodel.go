package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAttentionModel = (*customUserAttentionModel)(nil)

type (
	// UserAttentionModel 自定义接口
	UserAttentionModel interface {
		userAttentionModel
		FindOneByUserAndAttention(ctx context.Context, userId, attentionId int64) (*UserAttention, error)
		UpsertAttention(ctx context.Context, userId, attentionId int64) error
		SoftDeleteAttention(ctx context.Context, userId, attentionId int64) error
		ListAttentions(ctx context.Context, userId int64, offset, limit int64) ([]*UserAttention, int64, error)
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

// FindOneByUserAndAttention 根据 user_id + attention_id 查询（不使用缓存）
func (m *customUserAttentionModel) FindOneByUserAndAttention(ctx context.Context, userId, attentionId int64) (*UserAttention, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=? and `attention_id`=? limit 1", userAttentionRows, m.table)
	var resp UserAttention
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, attentionId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// UpsertAttention 关注：插入或恢复已删除的数据
func (m *customUserAttentionModel) UpsertAttention(ctx context.Context, userId, attentionId int64) error {
	// 先查
	record, err := m.FindOneByUserAndAttention(ctx, userId, attentionId)
	if err != nil && err != ErrNotFound {
		return err
	}
	if record == nil || err == ErrNotFound {
		query := fmt.Sprintf("insert into %s (`user_id`,`attention_id`,`create_time`,`is_del`) values (?, ?, ?, 0)", m.table)
		_, err := m.ExecNoCacheCtx(ctx, query, userId, attentionId, time.Now())
		return err
	}
	// 已存在且标记删除，则恢复
	if record.IsDel != 0 {
		query := fmt.Sprintf("update %s set `is_del`=0, `create_time`=? where `id`=?", m.table)
		_, err := m.ExecNoCacheCtx(ctx, query, time.Now(), record.Id)
		return err
	}
	// 已存在且正常，直接返回
	return nil
}

// SoftDeleteAttention 取关：将 is_del=1
func (m *customUserAttentionModel) SoftDeleteAttention(ctx context.Context, userId, attentionId int64) error {
	query := fmt.Sprintf("update %s set `is_del`=1 where `user_id`=? and `attention_id`=? and `is_del`=0", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, userId, attentionId)
	return err
}

// ListAttentions 关注列表（按创建时间倒序），返回列表和总数
func (m *customUserAttentionModel) ListAttentions(ctx context.Context, userId int64, offset, limit int64) ([]*UserAttention, int64, error) {
	listQuery := fmt.Sprintf("select %s from %s where `user_id`=? and `is_del`=0 order by `create_time` desc limit ? offset ?", userAttentionRows, m.table)
	var resp []*UserAttention
	if err := m.QueryRowsNoCacheCtx(ctx, &resp, listQuery, userId, limit, offset); err != nil && err != sql.ErrNoRows && err != sqlx.ErrNotFound {
		return nil, 0, err
	}
	countQuery := fmt.Sprintf("select count(1) from %s where `user_id`=? and `is_del`=0", m.table)
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, userId); err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}
