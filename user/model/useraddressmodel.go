package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAddressModel = (*customUserAddressModel)(nil)

type (
	// UserAddressModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAddressModel.
	UserAddressModel interface {
		userAddressModel

		// FindByUserId 查询用户所有地址
		FindByUserId(ctx context.Context, userId int64) ([]*UserAddress, error)
		// FindOneByUserIdAndAddressId 根据用户ID和地址ID查询地址（用于权限验证）
		FindOneByUserIdAndAddressId(ctx context.Context, userId int64, addressId uint64) (*UserAddress, error)
		// FindDefaultByUserId 查询用户的默认地址
		FindDefaultByUserId(ctx context.Context, userId int64) (*UserAddress, error)
		// ClearDefaultStatus 取消用户所有地址的默认状态
		ClearDefaultStatus(ctx context.Context, userId int64) error
		// Trans 事务支持
		Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error
	}

	customUserAddressModel struct {
		*defaultUserAddressModel
	}
)

// NewUserAddressModel returns a model for the database table.
func NewUserAddressModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserAddressModel {
	return &customUserAddressModel{
		defaultUserAddressModel: newUserAddressModel(conn, c, opts...),
	}
}

// FindByUserId 查询用户所有地址
func (m *customUserAddressModel) FindByUserId(ctx context.Context, userId int64) ([]*UserAddress, error) {
	var addresses []*UserAddress
	query := fmt.Sprintf("select %s from %s where user_id = ? order by default_status desc, create_time desc", userAddressRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &addresses, query, userId)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

// FindOneByUserIdAndAddressId 根据用户ID和地址ID查询地址（用于权限验证）
func (m *customUserAddressModel) FindOneByUserIdAndAddressId(ctx context.Context, userId int64, addressId uint64) (*UserAddress, error) {
	var address UserAddress
	query := fmt.Sprintf("select %s from %s where id = ? and user_id = ? limit 1", userAddressRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &address, query, addressId, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &address, nil
}

// FindDefaultByUserId 查询用户的默认地址
func (m *customUserAddressModel) FindDefaultByUserId(ctx context.Context, userId int64) (*UserAddress, error) {
	var address UserAddress
	query := fmt.Sprintf("select %s from %s where user_id = ? and default_status = 1 limit 1", userAddressRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &address, query, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &address, nil
}

// ClearDefaultStatus 取消用户所有地址的默认状态
func (m *customUserAddressModel) ClearDefaultStatus(ctx context.Context, userId int64) error {
	query := fmt.Sprintf("update %s set default_status = 0 where user_id = ? and default_status = 1", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, userId)
	return err
}

// Trans 事务支持
func (m *customUserAddressModel) Trans(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return m.TransactCtx(ctx, fn)
}
