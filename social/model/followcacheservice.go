package model

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// FollowCacheService 关注/粉丝列表缓存服务
// 封装了关注列表和粉丝列表的缓存逻辑
type FollowCacheService struct {
	userAttentionModel UserAttentionModel
	userFollowerModel  UserFollowerModel
	followListCache    *redisPool.RedisCache
	fansListCache      *redisPool.RedisCache
	followStatCache    *redisPool.RedisCache
}

// NewFollowCacheService 创建关注/粉丝列表缓存服务
// 如果 followListCache 或 fansListCache 为 nil，则创建新的 RedisCache 实例
func NewFollowCacheService(
	userAttentionModel UserAttentionModel,
	userFollowerModel UserFollowerModel,
	followListCache *redisPool.RedisCache,
	fansListCache *redisPool.RedisCache,
	followStatCache *redisPool.RedisCache,
) *FollowCacheService {
	if followListCache == nil {
		followListCache = redisPool.NewRedisCache("social", "follow_list")
	}
	if fansListCache == nil {
		fansListCache = redisPool.NewRedisCache("social", "fans_list")
	}
	if followStatCache == nil {
		followStatCache = redisPool.NewRedisCache("social", "follow_stat")
	}
	return &FollowCacheService{
		userAttentionModel: userAttentionModel,
		userFollowerModel:  userFollowerModel,
		followListCache:    followListCache,
		fansListCache:      fansListCache,
		followStatCache:    followStatCache,
	}
}

// GetFollowList 获取关注列表（带缓存）
// 优先从 Redis list 读取，如果没有则从数据库加载并写入缓存
func (s *FollowCacheService) GetFollowList(ctx context.Context, userId int64, page, size int32) ([]int64, int64, error) {
	// 优先使用 ZSet 实现（去重、按时间排序）
	result, total, err := s.GetFollowListFromZSet(ctx, userId, page, size)
	if err == nil && len(result) > 0 {
		return result, total, nil
	}

	// 如果 ZSet 中没有数据或发生错误，从数据库读取并写入 ZSet（重建）
	offset := (page - 1) * size
	attentions, totalCnt, err := s.userAttentionModel.ListAttentions(ctx, userId, int64(offset), int64(size))
	if err != nil {
		return nil, 0, err
	}
	if len(attentions) == 0 {
		return []int64{}, totalCnt, nil
	}

	// 重建 ZSet：先删除原 key，再按顺序 ZADD（保证 score 随索引下降，使得 ZREVRANGE 得到时间倒序）
	followListKey := fmt.Sprintf("follow_%d", userId)
	_ = s.followListCache.Delete(ctx, followListKey)
	base := time.Now().Unix()
	for i := 0; i < len(attentions); i++ {
		member := fmt.Sprintf("%d", attentions[i].AttentionId)
		score := base + int64(len(attentions)-i)
		_ = s.followListCache.ZAdd(ctx, followListKey, score, member)
	}

	// 读取并返回指定页
	return s.GetFollowListFromZSet(ctx, userId, page, size)
}

// GetFansList 获取粉丝列表（带缓存）
// 优先从 Redis list 读取，如果没有则从数据库加载并写入缓存
func (s *FollowCacheService) GetFansList(ctx context.Context, userId int64, page, size int32) ([]int64, int64, error) {
	// 优先使用 ZSet 实现
	result, total, err := s.GetFansListFromZSet(ctx, userId, page, size)
	if err == nil && len(result) > 0 {
		return result, total, nil
	}

	// 从数据库读取并写入 ZSet（重建）
	offset := (page - 1) * size
	followers, totalCnt, err := s.userFollowerModel.ListFollowers(ctx, userId, int64(offset), int64(size))
	if err != nil {
		return nil, 0, err
	}
	if len(followers) == 0 {
		return []int64{}, totalCnt, nil
	}

	fansListKey := fmt.Sprintf("fans_%d", userId)
	_ = s.fansListCache.Delete(ctx, fansListKey)
	base := time.Now().Unix()
	for i := 0; i < len(followers); i++ {
		member := fmt.Sprintf("%d", followers[i].FollowerId)
		score := base + int64(len(followers)-i)
		_ = s.fansListCache.ZAdd(ctx, fansListKey, score, member)
	}

	return s.GetFansListFromZSet(ctx, userId, page, size)
}

// AddToFollowList 添加用户到关注列表缓存（头插）
func (s *FollowCacheService) AddToFollowList(ctx context.Context, userId, targetId int64) error {
	followListKey := fmt.Sprintf("follow_%d", userId)
	member := fmt.Sprintf("%d", targetId)
	_, err := s.followListCache.Eval(ctx, luaAttentionsZset, []string{followListKey}, member)
	if err != nil {
		// 回退到 ZADD（兼容）
		if addErr := s.followListCache.ZAdd(ctx, followListKey, time.Now().Unix(), member); addErr != nil {
			return err
		}
		// 回退后补偿截断，保证长度不超过阈值
		if size, scErr := s.followListCache.ZCard(ctx, followListKey); scErr == nil {
			const maxLen = int64(2000)
			if size > maxLen {
				_, _ = s.followListCache.ZRemRangeByRank(ctx, followListKey, 0, size-maxLen-1)
			}
		}
		return err
	}
	return nil
}

// RemoveFromFollowList 从关注列表缓存中删除（删除整个列表，交给下次读重建）
func (s *FollowCacheService) RemoveFromFollowList(ctx context.Context, userId int64) error {
	followListKey := fmt.Sprintf("follow_%d", userId)
	return s.followListCache.Delete(ctx, followListKey)
}

// AddToFansList 添加用户到粉丝列表缓存（头插）
func (s *FollowCacheService) AddToFansList(ctx context.Context, userId, followerId int64) error {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	member := fmt.Sprintf("%d", followerId)
	_, err := s.fansListCache.Eval(ctx, luaFollowersZset, []string{fansListKey}, member)
	if err != nil {
		// 回退到 ZADD（兼容）
		if addErr := s.fansListCache.ZAdd(ctx, fansListKey, time.Now().Unix(), member); addErr != nil {
			return err
		}
		// 回退后补偿截断，保证长度不超过阈值
		if size, scErr := s.fansListCache.ZCard(ctx, fansListKey); scErr == nil {
			const maxLen = int64(10000)
			if size > maxLen {
				_, _ = s.fansListCache.ZRemRangeByRank(ctx, fansListKey, 0, size-maxLen-1)
			}
		}
		return err
	}
	return nil
}

// RemoveFromFansList 从粉丝列表缓存中删除（删除整个列表，交给下次读重建）
func (s *FollowCacheService) RemoveFromFansList(ctx context.Context, userId int64) error {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	return s.fansListCache.Delete(ctx, fansListKey)
}

// CheckFollowListContains 检查关注列表缓存中是否包含指定用户
func (s *FollowCacheService) CheckFollowListContains(ctx context.Context, userId, targetId int64) (bool, error) {
	followListKey := fmt.Sprintf("follow_%d", userId)
	member := fmt.Sprintf("%d", targetId)
	rank, err := s.followListCache.ZRank(ctx, followListKey, member)
	if err != nil {
		return false, err
	}
	return rank >= 0, nil
}

// CheckFansListContains 检查粉丝列表缓存中是否包含指定用户
func (s *FollowCacheService) CheckFansListContains(ctx context.Context, userId, followerId int64) (bool, error) {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	member := fmt.Sprintf("%d", followerId)
	rank, err := s.fansListCache.ZRank(ctx, fansListKey, member)
	if err != nil {
		return false, err
	}
	return rank >= 0, nil
}

// ---------------- ZSet 备用实现（可直接调用，不做自动切换） ----------------

// AddToFollowListZSet 使用 ZSET 插入关注（去重、按时间戳排序），脚本负责截断
func (s *FollowCacheService) AddToFollowListZSet(ctx context.Context, userId, targetId int64) error {
	followListKey := fmt.Sprintf("follow_%d", userId)
	member := fmt.Sprintf("%d", targetId)
	_, err := s.followListCache.Eval(ctx, luaAttentionsZset, []string{followListKey}, member)
	if err != nil {
		// 回退到 ZAdd（兼容）: 使用当前时间戳作为 score
		_ = s.followListCache.ZAdd(ctx, followListKey, time.Now().Unix(), member)
		return err
	}
	return nil
}

// AddToFansListZSet 使用 ZSET 插入粉丝（去重、按时间戳排序），脚本负责截断
func (s *FollowCacheService) AddToFansListZSet(ctx context.Context, userId, followerId int64) error {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	member := fmt.Sprintf("%d", followerId)
	_, err := s.fansListCache.Eval(ctx, luaFollowersZset, []string{fansListKey}, member)
	if err != nil {
		// 回退到 ZAdd（兼容）
		_ = s.fansListCache.ZAdd(ctx, fansListKey, time.Now().Unix(), member)
		return err
	}
	return nil
}

// RemoveFromFollowListZSet 从 ZSET 中删除指定 member
func (s *FollowCacheService) RemoveFromFollowListZSet(ctx context.Context, userId, targetId int64) (int64, error) {
	followListKey := fmt.Sprintf("follow_%d", userId)
	return s.followListCache.ZRem(ctx, followListKey, fmt.Sprintf("%d", targetId))
}

// RemoveFromFansListZSet 从 ZSET 中删除指定 member
func (s *FollowCacheService) RemoveFromFansListZSet(ctx context.Context, userId, followerId int64) (int64, error) {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	return s.fansListCache.ZRem(ctx, fansListKey, fmt.Sprintf("%d", followerId))
}

// GetFollowListFromZSet 使用 ZREVRANGE 分页读取 ZSET（按时间倒序）
func (s *FollowCacheService) GetFollowListFromZSet(ctx context.Context, userId int64, page, size int32) ([]int64, int64, error) {
	followListKey := fmt.Sprintf("follow_%d", userId)
	start := (page - 1) * size
	stop := start + size - 1
	values, err := s.followListCache.ZRevRange(ctx, followListKey, int64(start), int64(stop))
	if err != nil {
		return nil, 0, err
	}
	total, err := s.followListCache.ZCard(ctx, followListKey)
	if err != nil {
		return nil, 0, err
	}
	result := make([]int64, 0, len(values))
	for _, v := range values {
		n, err := parseInt64(v)
		if err != nil {
			continue
		}
		result = append(result, n)
	}
	return result, total, nil
}

// GetFansListFromZSet 使用 ZREVRANGE 分页读取 ZSET（按时间倒序）
func (s *FollowCacheService) GetFansListFromZSet(ctx context.Context, userId int64, page, size int32) ([]int64, int64, error) {
	fansListKey := fmt.Sprintf("fans_%d", userId)
	start := (page - 1) * size
	stop := start + size - 1
	values, err := s.fansListCache.ZRevRange(ctx, fansListKey, int64(start), int64(stop))
	if err != nil {
		return nil, 0, err
	}
	total, err := s.fansListCache.ZCard(ctx, fansListKey)
	if err != nil {
		return nil, 0, err
	}
	result := make([]int64, 0, len(values))
	for _, v := range values {
		n, err := parseInt64(v)
		if err != nil {
			continue
		}
		result = append(result, n)
	}
	return result, total, nil
}

// parseInt64 解析字符串为 int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
