package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		// 扩展方法：带缓存的用户服务
		CacheService() *UserCacheService
	}

	customUserModel struct {
		*defaultUserModel
		cacheService *UserCacheService
	}
)

// NewUserModel returns a model for the database table.
// 如果 userCache 为 nil，则会在 NewUserCacheService 中创建新的实例
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, userCache *redisPool.RedisCache, opts ...cache.Option) UserModel {
	model := &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
	// 初始化缓存服务
	model.cacheService = NewUserCacheService(model, userCache)
	return model
}

// CacheService 获取用户缓存服务
func (m *customUserModel) CacheService() *UserCacheService {
	return m.cacheService
}

// GetUserWithCache 根据ID获取用户信息（使用缓存）
// 这是一个便捷方法，直接调用缓存服务
func (m *customUserModel) GetUserWithCache(ctx context.Context, userId uint64) (*UserDTO, error) {
	return m.cacheService.GetUserById(ctx, userId)
}
