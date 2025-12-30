package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// FollowModel 关注/粉丝相关业务模型
// 封装了关注列表和粉丝列表的缓存逻辑
type FollowModel struct {
	userAttentionModel UserAttentionModel
	userFollowerModel  UserFollowerModel
	userRelationModel  UserRelationModel
	cacheService       *FollowCacheService
}

// NewFollowModel 创建关注/粉丝业务模型
// 如果 followListCache 或 fansListCache 为 nil，则会在 NewFollowCacheService 中创建新的实例
func NewFollowModel(
	conn sqlx.SqlConn,
	c cache.CacheConf,
	followListCache *redisPool.RedisCache,
	fansListCache *redisPool.RedisCache,
	followStatCache *redisPool.RedisCache,
	opts ...cache.Option,
) *FollowModel {
	// 初始化基础模型
	userAttentionModel := NewUserAttentionModel(conn, c, opts...)
	userFollowerModel := NewUserFollowerModel(conn, c, opts...)
	userRelationModel := NewUserRelationModel(conn, c, opts...)

	// 初始化缓存服务
	cacheService := NewFollowCacheService(
		userAttentionModel,
		userFollowerModel,
		followListCache,
		fansListCache,
		followStatCache,
	)

	return &FollowModel{
		userAttentionModel: userAttentionModel,
		userFollowerModel:  userFollowerModel,
		userRelationModel:  userRelationModel,
		cacheService:       cacheService,
	}
}

// UserAttentionModel 获取用户关注模型
func (m *FollowModel) UserAttentionModel() UserAttentionModel {
	return m.userAttentionModel
}

// UserFollowerModel 获取用户粉丝模型
func (m *FollowModel) UserFollowerModel() UserFollowerModel {
	return m.userFollowerModel
}

// UserRelationModel 获取用户关系模型
func (m *FollowModel) UserRelationModel() UserRelationModel {
	return m.userRelationModel
}

// CacheService 获取关注/粉丝列表缓存服务
func (m *FollowModel) CacheService() *FollowCacheService {
	return m.cacheService
}
