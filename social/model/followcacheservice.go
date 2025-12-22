package model

import (
	"context"
	"fmt"
	"strconv"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// FollowCacheService 关注/粉丝列表缓存服务
// 封装了关注列表和粉丝列表的缓存逻辑
type FollowCacheService struct {
	userAttentionModel UserAttentionModel
	userFollowerModel  UserFollowerModel
	followListCache    *redisPool.RedisCache
	fansListCache      *redisPool.RedisCache
}

// NewFollowCacheService 创建关注/粉丝列表缓存服务
// 如果 followListCache 或 fansListCache 为 nil，则创建新的 RedisCache 实例
func NewFollowCacheService(
	userAttentionModel UserAttentionModel,
	userFollowerModel UserFollowerModel,
	followListCache *redisPool.RedisCache,
	fansListCache *redisPool.RedisCache,
) *FollowCacheService {
	if followListCache == nil {
		followListCache = redisPool.NewRedisCache("social", "follow_list")
	}
	if fansListCache == nil {
		fansListCache = redisPool.NewRedisCache("social", "fans_list")
	}
	return &FollowCacheService{
		userAttentionModel: userAttentionModel,
		userFollowerModel:  userFollowerModel,
		followListCache:    followListCache,
		fansListCache:      fansListCache,
	}
}

// GetFollowList 获取关注列表（带缓存）
// 优先从 Redis list 读取，如果没有则从数据库加载并写入缓存
func (s *FollowCacheService) GetFollowList(ctx context.Context, userId int64, page, size int32) ([]int64, int64, error) {
	followListKey := fmt.Sprintf("follow_%d", userId)

	// 计算分页范围
	start := (page - 1) * size
	stop := start + size - 1

	// 从 Redis list 读取
	userIds, err := s.followListCache.LRange(ctx, followListKey, int64(start), int64(stop))
	if err != nil {
		// 忽略错误，继续从数据库查询
	}

	// 如果 Redis 中没有数据，从数据库读取并写入 Redis
	if len(userIds) == 0 {
		offset := (page - 1) * size
		attentions, total, err := s.userAttentionModel.ListAttentions(ctx, userId, int64(offset), int64(size))
		if err != nil {
			return nil, 0, err
		}

		// 写入 Redis list（从左侧插入，保持时间倒序）
		if total > 0 {
			// 先清空列表（如果存在），然后重新构建
			_ = s.followListCache.Delete(ctx, followListKey)
			for i := total - 1; i >= 0; i-- {
				attentionIdStr := fmt.Sprintf("%d", attentions[i].AttentionId)
				_, _ = s.followListCache.LPush(ctx, followListKey, attentionIdStr)
			}
			// 重新读取当前页
			userIds, _ = s.followListCache.LRange(ctx, followListKey, int64(start), int64(stop))
		} else {
			// 没有数据，返回空列表
			return []int64{}, 0, nil
		}
	}

	// 获取总数（从 Redis list 长度或数据库）
	total, err := s.followListCache.LLen(ctx, followListKey)
	if err != nil || total == 0 {
		// 如果 Redis 中没有总数，从数据库查询
		_, total, err = s.userAttentionModel.ListAttentions(ctx, userId, 0, 1)
		if err != nil {
			total = int64(len(userIds))
		}
	}

	// 转换为 int64 切片
	result := make([]int64, 0, len(userIds))
	for _, userIdStr := range userIds {
		attentionId, err := parseInt64(userIdStr)
		if err != nil {
			continue
		}
		result = append(result, attentionId)
	}

	return result, total, nil
}

// GetFansList 获取粉丝列表（带缓存）
// 优先从 Redis list 读取，如果没有则从数据库加载并写入缓存
func (s *FollowCacheService) GetFansList(ctx context.Context, userId int64, page, size int32) ([]int64, int64, error) {
	fansListKey := fmt.Sprintf("fans_%d", userId)

	// 计算分页范围
	start := (page - 1) * size
	stop := start + size - 1

	// 从 Redis list 读取
	userIds, err := s.fansListCache.LRange(ctx, fansListKey, int64(start), int64(stop))
	if err != nil {
		// 忽略错误，继续从数据库查询
	}

	// 如果 Redis 中没有数据，从数据库读取并写入 Redis
	if len(userIds) == 0 {
		offset := (page - 1) * size
		followers, total, err := s.userFollowerModel.ListFollowers(ctx, userId, int64(offset), int64(size))
		if err != nil {
			return nil, 0, err
		}

		// 写入 Redis list（从左侧插入，保持时间倒序）
		if len(followers) > 0 {
			// 先清空列表（如果存在），然后重新构建
			_ = s.fansListCache.Delete(ctx, fansListKey)
			for i := len(followers) - 1; i >= 0; i-- {
				followerIdStr := fmt.Sprintf("%d", followers[i].FollowerId)
				_, _ = s.fansListCache.LPush(ctx, fansListKey, followerIdStr)
			}
			// 重新读取当前页
			userIds, _ = s.fansListCache.LRange(ctx, fansListKey, int64(start), int64(stop))
		} else {
			// 没有数据，返回空列表
			return []int64{}, total, nil
		}
	}

	// 获取总数（从 Redis list 长度或数据库）
	total, err := s.fansListCache.LLen(ctx, fansListKey)
	if err != nil || total == 0 {
		// 如果 Redis 中没有总数，从数据库查询
		_, total, err = s.userFollowerModel.ListFollowers(ctx, userId, 0, 1)
		if err != nil {
			total = int64(len(userIds))
		}
	}

	// 转换为 int64 切片
	result := make([]int64, 0, len(userIds))
	for _, userIdStr := range userIds {
		followerId, err := parseInt64(userIdStr)
		if err != nil {
			continue
		}
		result = append(result, followerId)
	}

	return result, total, nil
}

// AddToFollowList 添加用户到关注列表缓存（头插）
func (s *FollowCacheService) AddToFollowList(ctx context.Context, userId, targetId int64) error {
	followListKey := fmt.Sprintf("follow_%d", userId)
	targetIdStr := fmt.Sprintf("%d", targetId)
	_, err := s.followListCache.LPush(ctx, followListKey, targetIdStr)
	return err
}

// RemoveFromFollowList 从关注列表缓存中删除（删除整个列表，交给下次读重建）
func (s *FollowCacheService) RemoveFromFollowList(ctx context.Context, userId int64) error {
	followListKey := fmt.Sprintf("follow_%d", userId)
	return s.followListCache.Delete(ctx, followListKey)
}

// AddToFansList 添加用户到粉丝列表缓存（头插）
func (s *FollowCacheService) AddToFansList(ctx context.Context, userId, followerId int64) error {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	followerIdStr := fmt.Sprintf("%d", followerId)
	_, err := s.fansListCache.LPush(ctx, fansListKey, followerIdStr)
	return err
}

// RemoveFromFansList 从粉丝列表缓存中删除（删除整个列表，交给下次读重建）
func (s *FollowCacheService) RemoveFromFansList(ctx context.Context, userId int64) error {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	return s.fansListCache.Delete(ctx, fansListKey)
}

// CheckFollowListContains 检查关注列表缓存中是否包含指定用户
func (s *FollowCacheService) CheckFollowListContains(ctx context.Context, userId, targetId int64) (bool, error) {
	followListKey := fmt.Sprintf("follow_%d", userId)
	targetIdStr := fmt.Sprintf("%d", targetId)
	values, err := s.followListCache.LRange(ctx, followListKey, 0, -1)
	if err != nil {
		return false, err
	}
	for _, v := range values {
		if v == targetIdStr {
			return true, nil
		}
	}
	return false, nil
}

// CheckFansListContains 检查粉丝列表缓存中是否包含指定用户
func (s *FollowCacheService) CheckFansListContains(ctx context.Context, userId, followerId int64) (bool, error) {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	followerIdStr := fmt.Sprintf("%d", followerId)
	values, err := s.fansListCache.LRange(ctx, fansListKey, 0, -1)
	if err != nil {
		return false, err
	}
	for _, v := range values {
		if v == followerIdStr {
			return true, nil
		}
	}
	return false, nil
}

// parseInt64 解析字符串为 int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

