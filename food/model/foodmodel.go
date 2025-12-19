package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

var _ FoodModel = (*customFoodModel)(nil)

type (
	// FoodModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFoodModel.
	FoodModel interface {
		foodModel
		// 查询我的美食列表（根据 user_id，支持分页）
		FindListByUserId(ctx context.Context, userId int64, page, pageSize int32) ([]*Food, int64, error)
		// 扩展方法：带缓存的美食服务
		CacheService() *FoodCacheService
	}

	customFoodModel struct {
		*defaultFoodModel
		cacheService *FoodCacheService
	}
)

// NewFoodModel returns a model for the database table.
// 如果 foodCache 为 nil，则会在 NewFoodCacheService 中创建新的实例
func NewFoodModel(conn sqlx.SqlConn, c cache.CacheConf, foodCache *redisPool.RedisCache, opts ...cache.Option) FoodModel {
	model := &customFoodModel{
		defaultFoodModel: newFoodModel(conn, c, opts...),
	}
	// 初始化缓存服务
	model.cacheService = NewFoodCacheService(model, foodCache)
	return model
}

// CacheService 获取美食缓存服务
func (m *customFoodModel) CacheService() *FoodCacheService {
	return m.cacheService
}

// FindListByUserId 查询我的美食列表（根据 user_id 和分类，支持分页）
func (m *customFoodModel) FindListByUserId(ctx context.Context, userId int64, page, pageSize int32) ([]*Food, int64, error) {
	// 构建查询条件
	where := "`user_id` = ? AND `food_status` = 0"
	args := []interface{}{userId}

	// 查询总数
	var total int64
	countQuery := "SELECT COUNT(*) FROM `food` WHERE " + where
	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 如果没有数据，直接返回
	if total == 0 {
		return []*Food{}, 0, nil
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY `food_createtime` DESC LIMIT ? OFFSET ?", foodRows, m.table, where)
	args = append(args, pageSize, offset)

	var foods []*Food
	err = m.QueryRowsNoCacheCtx(ctx, &foods, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*Food{}, total, nil
		}
		return nil, 0, err
	}

	return foods, total, nil
}
