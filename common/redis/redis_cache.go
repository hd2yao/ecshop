package redis_pool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/hd2yao/ecshop/common/errcode"
)

// ==================== 常量定义 ====================

const (
	// EmptyCacheValue 空值标记，用于防止缓存穿透
	EmptyCacheValue = "EMPTY_CACHE_MARKER"
	// DefaultEmptyExpiry 空值默认过期时间
	DefaultEmptyExpiry = 5 * time.Minute
	// DefaultRandomExpiry 随机过期时间范围（防止缓存雪崩）
	DefaultRandomExpiry = 24 * time.Hour
	// DefaultBaseExpiry 基础过期时间
	DefaultBaseExpiry = 48 * time.Hour
	// DefaultAutoRefreshThreshold 自动续期阈值（剩余过期时间百分比）
	DefaultAutoRefreshThreshold = 0.2
)

// ==================== 类型定义 ====================

// RedisCache Redis 缓存工具类
// 提供两大类功能：
// 1. 智能缓存：防穿透、防击穿、自动续期、强一致性
// 2. 基础操作：String、Set、ZSet、Hash、List 等 Redis 数据结构操作
type RedisCache struct {
	client       *redis.Redis
	keyBuilder   *RedisKeyBuilder
	lockExecutor *LockExecutor
}

// CacheOption 缓存配置选项
type CacheOption struct {
	BaseExpiry           time.Duration // 基础过期时间
	RandomExpiry         time.Duration // 随机过期时间范围（防止缓存雪崩）
	EmptyExpiry          time.Duration // 空值过期时间（防止缓存穿透）
	EnableAutoRefresh    bool          // 是否启用自动续期
	AutoRefreshThreshold float64       // 自动续期阈值（剩余过期时间百分比）
}

// DefaultCacheOption 默认缓存配置
func DefaultCacheOption() *CacheOption {
	return &CacheOption{
		BaseExpiry:           DefaultBaseExpiry,
		RandomExpiry:         DefaultRandomExpiry,
		EmptyExpiry:          DefaultEmptyExpiry,
		EnableAutoRefresh:    true,
		AutoRefreshThreshold: DefaultAutoRefreshThreshold,
	}
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache(module, business string) *RedisCache {
	return &RedisCache{
		client:       GetRedisClient(),
		keyBuilder:   NewRedisKeyBuilder(module, business),
		lockExecutor: NewLockExecutor(),
	}
}

// ==================== 核心智能缓存方法 ====================

// GetWithLoader 获取缓存，如果不存在则通过 loader 函数加载
//
// 这是最核心的方法，集成了以下高级特性：
//   - 缓存穿透防护：空值缓存
//   - 缓存击穿防护：分布式锁 + Double Check
//   - 缓存雪崩防护：随机过期时间
//   - 热点数据自动延期：访问频繁的数据自动续期
//   - 高并发优化：串行转并发，未获取锁的线程自动降级
//
// 适用场景：高并发读场景（如首页 Feed 流、商品详情页）
func (c *RedisCache) GetWithLoader(
	ctx context.Context,
	key string,
	result interface{},
	loader func() (interface{}, error),
	opt *CacheOption,
) error {
	if opt == nil {
		opt = DefaultCacheOption()
	}

	cacheKey := c.keyBuilder.BuildKey(key)

	// 第一次尝试：直接从缓存读取
	ttl, err := c.getFromCache(ctx, cacheKey, result)
	if err == nil {
		// 缓存命中，检查是否需要自动续期
		if opt.EnableAutoRefresh {
			c.tryAutoRefresh(ctx, cacheKey, ttl, opt)
		}
		return nil
	}

	// 如果是空值缓存，返回错误
	if errors.Is(err, errcode.CacheEmpty) {
		return errcode.CacheNotFound
	}

	// 缓存未命中，需要从数据源加载
	// 使用分布式锁保护，防止缓存击穿
	lockKey := c.keyBuilder.BuildKey(key + ":lock")

	// 尝试获取锁（非阻塞）
	executed, err := c.lockExecutor.ExecuteWithTryLock(ctx, lockKey, 10*time.Second, func() error {
		// Double Check：再次检查缓存，避免重复加载
		_, err := c.getFromCache(ctx, cacheKey, result)
		if err == nil {
			return nil
		}

		// 从数据源加载
		data, loadErr := loader()
		if loadErr != nil {
			return fmt.Errorf("加载数据失败: %w", loadErr)
		}

		// 如果数据为空，设置空值缓存防止穿透
		if data == nil {
			return c.SetEmpty(ctx, key, opt.EmptyExpiry)
		}

		// 设置缓存
		expiry := c.calculateExpiry(opt)
		return c.Set(ctx, key, data, expiry)
	})

	if err != nil {
		return err
	}

	// 如果没有执行（说明没获取到锁）
	if !executed {
		// 短暂等待后重试读取缓存（此时其他线程可能已经加载完成）
		time.Sleep(50 * time.Millisecond)

		// 再次尝试从缓存读取
		_, err := c.getFromCache(ctx, cacheKey, result)
		if err == nil {
			return nil
		}

		// 如果还是读不到，可能是并发量太大，直接穿透到数据库
		// 这里不再使用阻塞锁，避免大量线程排队
		data, loadErr := loader()
		if loadErr != nil {
			return fmt.Errorf("加载数据失败: %w", loadErr)
		}

		if data == nil {
			return errcode.CacheNotFound
		}

		// 将数据赋值给 result
		return c.unmarshalData(data, result)
	}

	// 重新读取缓存
	_, err = c.getFromCache(ctx, cacheKey, result)
	return err
}

// GetWithLoaderAndMutex 获取缓存，使用互斥锁保证读写串行化（强一致性）
//
// 使用场景：需要保证数据库和缓存强一致性的场景（如个人中心、账户设置）
//
// 与 GetWithLoader 的区别：
//   - GetWithLoader: 非阻塞锁，高并发优化，允许短暂不一致
//   - GetWithLoaderAndMutex: 互斥锁，读写串行化，保证强一致性
func (c *RedisCache) GetWithLoaderAndMutex(
	ctx context.Context,
	key string,
	result interface{},
	loader func() (interface{}, error),
	opt *CacheOption,
) error {
	if opt == nil {
		opt = DefaultCacheOption()
	}

	cacheKey := c.keyBuilder.BuildKey(key)
	lockKey := c.keyBuilder.BuildKey(key + ":mutex")

	// 使用分布式锁保证读写互斥
	return c.lockExecutor.ExecuteWithLock(ctx, lockKey, 10*time.Second, 3*time.Second, func() error {
		// Double Check：在锁内再次检查缓存
		ttl, err := c.getFromCache(ctx, cacheKey, result)
		if err == nil {
			// 缓存命中，检查是否需要自动续期
			if opt.EnableAutoRefresh {
				c.tryAutoRefresh(ctx, cacheKey, ttl, opt)
			}
			return nil
		}

		// 如果是空值缓存，返回错误
		if errors.Is(err, errcode.CacheEmpty) {
			return errcode.CacheNotFound
		}

		// 从数据源加载
		data, loadErr := loader()
		if loadErr != nil {
			return fmt.Errorf("加载数据失败: %w", loadErr)
		}

		// 如果数据为空，设置空值缓存防止穿透
		if data == nil {
			return c.SetEmpty(ctx, key, opt.EmptyExpiry)
		}

		// 设置缓存
		expiry := c.calculateExpiry(opt)
		if err := c.Set(ctx, key, data, expiry); err != nil {
			return err
		}

		// 将数据赋值给 result
		return c.unmarshalData(data, result)
	})
}

// UpdateWithMutex 在分布式锁保护下更新缓存和数据库
//
// 用于写操作，保证缓存和数据库的强一致性
// 使用与 GetWithLoaderAndMutex 相同的互斥锁，保证读写互斥
//
// 使用场景：更新用户信息、修改订单状态等需要强一致性的写操作
func (c *RedisCache) UpdateWithMutex(
	ctx context.Context,
	key string,
	updater func() error,
	opt *CacheOption,
) error {
	if opt == nil {
		opt = DefaultCacheOption()
	}

	lockKey := c.keyBuilder.BuildKey(key + ":mutex")

	// 使用分布式锁保证读写互斥
	return c.lockExecutor.ExecuteWithLock(ctx, lockKey, 10*time.Second, 3*time.Second, func() error {
		// 执行更新操作（更新数据库）
		if err := updater(); err != nil {
			return fmt.Errorf("更新数据失败: %w", err)
		}

		// 删除缓存，让下次读取时重新加载最新数据
		cacheKey := c.keyBuilder.BuildKey(key)
		_, _ = c.client.DelCtx(ctx, cacheKey)

		return nil
	})
}

// ==================== 基础缓存操作 ====================

// Set 设置缓存（带过期时间）
//
// 自动将 value 序列化为 JSON 字符串存储
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	cacheKey := c.keyBuilder.BuildKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	return c.client.SetexCtx(ctx, cacheKey, string(data), int(expiry.Seconds()))
}

// SetEmpty 设置空值缓存（防止缓存穿透）
//
// 当查询数据不存在时，设置一个特殊标记，避免重复查询数据库
func (c *RedisCache) SetEmpty(ctx context.Context, key string, expiry time.Duration) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	return c.client.SetexCtx(ctx, cacheKey, EmptyCacheValue, int(expiry.Seconds()))
}

// Get 获取缓存
//
// 自动将 JSON 字符串反序列化到 result 对象
func (c *RedisCache) Get(ctx context.Context, key string, result interface{}) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	_, err := c.getFromCache(ctx, cacheKey, result)
	return err
}

// Delete 删除单个缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	_, err := c.client.DelCtx(ctx, cacheKey)
	return err
}

// DeleteBatch 批量删除缓存
func (c *RedisCache) DeleteBatch(ctx context.Context, keys ...string) error {
	cacheKeys := c.keyBuilder.BuildKeys(keys...)
	_, err := c.client.DelCtx(ctx, cacheKeys...)
	return err
}

// Refresh 刷新缓存过期时间
func (c *RedisCache) Refresh(ctx context.Context, key string, expiry time.Duration) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	err := c.client.ExpireCtx(ctx, cacheKey, int(expiry.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

// Exists 判断 key 是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	exists, err := c.client.ExistsCtx(ctx, cacheKey)
	if err != nil {
		return false, fmt.Errorf("判断key存在失败: %w", err)
	}
	return exists, nil
}

// GetExpire 获取 key 的过期时间（秒）
//
// 返回值：
//   - > 0: 剩余过期时间（秒）
//   - -1: key 存在但没有设置过期时间
//   - -2: key 不存在
func (c *RedisCache) GetExpire(ctx context.Context, key string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	ttl, err := c.client.TtlCtx(ctx, cacheKey)
	if err != nil {
		return 0, fmt.Errorf("获取过期时间失败: %w", err)
	}
	return int64(ttl), nil
}

// ==================== String 操作 ====================

// MGet 批量获取多个 key 的值
//
// 使用场景：批量获取用户信息、商品信息等
func (c *RedisCache) MGet(ctx context.Context, keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return []string{}, nil
	}

	cacheKeys := c.keyBuilder.BuildKeys(keys...)
	values, err := c.client.MgetCtx(ctx, cacheKeys...)
	if err != nil {
		return nil, fmt.Errorf("批量获取缓存失败: %w", err)
	}

	return values, nil
}

// Increment 对 key 的值进行自增操作
//
// 使用场景：计数器、浏览次数、点赞数等
func (c *RedisCache) Increment(ctx context.Context, key string, increment int64) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	val, err := c.client.IncrbyCtx(ctx, cacheKey, increment)
	if err != nil {
		return 0, fmt.Errorf("自增操作失败: %w", err)
	}
	return val, nil
}

// Decrement 对 key 的值进行自减操作
//
// 使用场景：库存扣减、剩余次数等
func (c *RedisCache) Decrement(ctx context.Context, key string, decrement int64) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	val, err := c.client.DecrbyCtx(ctx, cacheKey, decrement)
	if err != nil {
		return 0, fmt.Errorf("自减操作失败: %w", err)
	}
	return val, nil
}

// ==================== Set 操作 ====================

// SAdd 向 Set 集合添加成员
//
// 使用场景：用户标签、关注列表、点赞用户列表等
// 返回值：成功添加的成员数量
func (c *RedisCache) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)

	// 将 []string 转换为 []interface{}
	values := make([]interface{}, len(members))
	for i, member := range members {
		values[i] = member
	}

	count, err := c.client.SaddCtx(ctx, cacheKey, values...)
	if err != nil {
		return 0, fmt.Errorf("添加Set成员失败: %w", err)
	}
	return int64(count), nil
}

// SRem 从 Set 集合删除成员
//
// 返回值：成功删除的成员数量
func (c *RedisCache) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)

	// 将 []string 转换为 []interface{}
	values := make([]interface{}, len(members))
	for i, member := range members {
		values[i] = member
	}

	count, err := c.client.SremCtx(ctx, cacheKey, values...)
	if err != nil {
		return 0, fmt.Errorf("删除Set成员失败: %w", err)
	}
	return int64(count), nil
}

// SMembers 获取 Set 集合所有成员
func (c *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	members, err := c.client.SmembersCtx(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("获取Set成员失败: %w", err)
	}
	return members, nil
}

// SIsMember 判断成员是否在 Set 集合中
//
// 使用场景：判断用户是否已点赞、是否已关注等
func (c *RedisCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	exists, err := c.client.SismemberCtx(ctx, cacheKey, member)
	if err != nil {
		return false, fmt.Errorf("判断Set成员存在失败: %w", err)
	}
	return exists, nil
}

// SCard 获取 Set 集合成员数量
//
// 使用场景：获取关注数、粉丝数、点赞数等
func (c *RedisCache) SCard(ctx context.Context, key string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	count, err := c.client.ScardCtx(ctx, cacheKey)
	if err != nil {
		return 0, fmt.Errorf("获取Set成员数量失败: %w", err)
	}
	return count, nil
}

// ==================== ZSet 操作（有序集合）====================

// ZAdd 向有序集合添加成员（整数分数）
//
// 使用场景：排行榜、热度排序等（分数为整数时使用）
func (c *RedisCache) ZAdd(ctx context.Context, key string, score int64, member string) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	_, err := c.client.ZaddCtx(ctx, cacheKey, score, member)
	if err != nil {
		return fmt.Errorf("添加ZSet成员失败: %w", err)
	}
	return nil
}

// ZAddFloat 向有序集合添加成员（浮点数分数）
//
// 使用场景：需要精确分数的排行榜（如热度算法计算的分数）
func (c *RedisCache) ZAddFloat(ctx context.Context, key string, score float64, member string) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	_, err := c.client.ZaddFloatCtx(ctx, cacheKey, score, member)
	if err != nil {
		return fmt.Errorf("添加ZSet成员失败: %w", err)
	}
	return nil
}

// ZRem 从有序集合删除成员
//
// 返回值：成功删除的成员数量
func (c *RedisCache) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	count, err := c.client.ZremCtx(ctx, cacheKey, members)
	if err != nil {
		return 0, fmt.Errorf("删除ZSet成员失败: %w", err)
	}
	return int64(count), nil
}

// ZRange 获取有序集合指定范围的成员（按索引，正序）
//
// 使用场景：获取排行榜指定名次范围的成员
//
// 参数：
//   - start: 起始索引（0表示第一个）
//   - stop: 结束索引（-1表示最后一个）
//
// 示例：ZRange(ctx, "rank", 0, 9) 获取前10名
func (c *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	members, err := c.client.ZrangeCtx(ctx, cacheKey, start, stop)
	if err != nil {
		return nil, fmt.Errorf("获取ZSet范围成员失败: %w", err)
	}
	return members, nil
}

// ZRevRange 获取有序集合指定范围的成员（按索引，倒序）
//
// 使用场景：获取排行榜（分数从高到低）
//
// 示例：ZRevRange(ctx, "rank", 0, 9) 获取排行榜前10名（分数最高的）
func (c *RedisCache) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	members, err := c.client.ZrevrangeCtx(ctx, cacheKey, start, stop)
	if err != nil {
		return nil, fmt.Errorf("倒序获取ZSet范围成员失败: %w", err)
	}
	return members, nil
}

// ZRangeByScore 按分数范围获取有序集合成员
//
// 使用场景：获取分数在某个范围内的成员（如获取80-100分的学生）
//
// 参数：
//   - min: 最小分数
//   - max: 最大分数
//   - offset: 偏移量（分页用）
//   - count: 返回数量（分页用）
func (c *RedisCache) ZRangeByScore(ctx context.Context, key string, min, max int64, offset, count int64) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)

	pairs, err := c.client.ZrangebyscoreWithScoresAndLimitCtx(ctx, cacheKey, min, max, int(offset), int(count))
	if err != nil {
		return nil, fmt.Errorf("按分数范围获取ZSet成员失败: %w", err)
	}

	// 提取成员（忽略分数）
	members := make([]string, len(pairs))
	for i, pair := range pairs {
		members[i] = pair.Key
	}

	return members, nil
}

// ZCard 获取有序集合成员数量
//
// 使用场景：获取排行榜总人数
func (c *RedisCache) ZCard(ctx context.Context, key string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	count, err := c.client.ZcardCtx(ctx, cacheKey)
	if err != nil {
		return 0, fmt.Errorf("获取ZSet成员数量失败: %w", err)
	}
	return int64(count), nil
}

// ==================== Hash 操作 ====================

// HSet 设置 Hash 单个字段的值
//
// 使用场景：存储用户信息、商品属性等结构化数据
func (c *RedisCache) HSet(ctx context.Context, key, field, value string) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	return c.client.HsetCtx(ctx, cacheKey, field, value)
}

// HGet 获取 Hash 单个字段的值
func (c *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	value, err := c.client.HgetCtx(ctx, cacheKey, field)
	if err != nil {
		return "", fmt.Errorf("获取Hash字段失败: %w", err)
	}
	return value, nil
}

// HMSet 批量设置 Hash 字段
//
// 使用场景：一次性设置多个字段值
func (c *RedisCache) HMSet(ctx context.Context, key string, fields map[string]string) error {
	cacheKey := c.keyBuilder.BuildKey(key)
	return c.client.HmsetCtx(ctx, cacheKey, fields)
}

// HMGet 批量获取 Hash 多个字段的值
//
// 使用场景：一次性获取多个字段值
func (c *RedisCache) HMGet(ctx context.Context, key string, fields ...string) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	values, err := c.client.HmgetCtx(ctx, cacheKey, fields...)
	if err != nil {
		return nil, fmt.Errorf("批量获取Hash字段失败: %w", err)
	}
	return values, nil
}

// HGetAll 获取 Hash 所有字段和值
// 使用场景：获取完整的用户信息、商品所有属性等
func (c *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	values, err := c.client.HgetallCtx(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("获取Hash所有字段失败: %w", err)
	}
	return values, nil
}

// HDel 删除 Hash 字段
func (c *RedisCache) HDel(ctx context.Context, key string, fields ...string) (bool, error) {
	cacheKey := c.keyBuilder.BuildKey(key)

	isDel, err := c.client.HdelCtx(ctx, cacheKey, fields...)
	if err != nil {
		return isDel, fmt.Errorf("删除Hash字段失败: %w", err)
	}
	return isDel, nil
}

// HExists 判断 Hash 字段是否存在
func (c *RedisCache) HExists(ctx context.Context, key, field string) (bool, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	exists, err := c.client.HexistsCtx(ctx, cacheKey, field)
	if err != nil {
		return false, fmt.Errorf("判断Hash字段存在失败: %w", err)
	}
	return exists, nil
}

// HKeys 获取 Hash 所有字段名
func (c *RedisCache) HKeys(ctx context.Context, key string) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	keys, err := c.client.HkeysCtx(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("获取Hash所有字段名失败: %w", err)
	}
	return keys, nil
}

// HVals 获取 Hash 所有字段值
func (c *RedisCache) HVals(ctx context.Context, key string) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	values, err := c.client.HvalsCtx(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("获取Hash所有字段值失败: %w", err)
	}
	return values, nil
}

// HLen 获取 Hash 字段数量
func (c *RedisCache) HLen(ctx context.Context, key string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	count, err := c.client.HlenCtx(ctx, cacheKey)
	if err != nil {
		return 0, fmt.Errorf("获取Hash字段数量失败: %w", err)
	}
	return int64(count), nil
}

// HIncrBy Hash 字段值自增
//
// 使用场景：用户积分增加、统计计数等
func (c *RedisCache) HIncrBy(ctx context.Context, key, field string, increment int64) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	val, err := c.client.HincrbyCtx(ctx, cacheKey, field, int(increment))
	if err != nil {
		return 0, fmt.Errorf("Hash字段自增失败: %w", err)
	}
	return int64(val), nil
}

// ==================== List 操作 ====================

// LPush 从列表左侧添加元素
//
// 使用场景：消息队列、最新动态列表等
// 返回值：添加后列表的长度
func (c *RedisCache) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)

	// 将 []string 转换为 []interface{}
	elements := make([]interface{}, len(values))
	for i, value := range values {
		elements[i] = value
	}

	count, err := c.client.LpushCtx(ctx, cacheKey, elements...)
	if err != nil {
		return 0, fmt.Errorf("从左侧添加List元素失败: %w", err)
	}
	return int64(count), nil
}

// RPush 从列表右侧添加元素
//
// 使用场景：消息队列（生产者）、时间线等
// 返回值：添加后列表的长度
func (c *RedisCache) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)

	// 将 []string 转换为 []interface{}
	elements := make([]interface{}, len(values))
	for i, value := range values {
		elements[i] = value
	}

	count, err := c.client.RpushCtx(ctx, cacheKey, elements...)
	if err != nil {
		return 0, fmt.Errorf("从右侧添加List元素失败: %w", err)
	}
	return int64(count), nil
}

// LPop 从列表左侧弹出元素
//
// 使用场景：消息队列（消费者）
func (c *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	value, err := c.client.LpopCtx(ctx, cacheKey)
	if err != nil {
		return "", fmt.Errorf("从左侧弹出List元素失败: %w", err)
	}
	return value, nil
}

// RPop 从列表右侧弹出元素
func (c *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	value, err := c.client.RpopCtx(ctx, cacheKey)
	if err != nil {
		return "", fmt.Errorf("从右侧弹出List元素失败: %w", err)
	}
	return value, nil
}

// LRange 获取列表指定范围的元素
//
// 参数：
//   - start: 起始索引（0表示第一个）
//   - stop: 结束索引（-1表示最后一个）
//
// 示例：LRange(ctx, "list", 0, 9) 获取前10个元素
func (c *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	values, err := c.client.LrangeCtx(ctx, cacheKey, int(start), int(stop))
	if err != nil {
		return nil, fmt.Errorf("获取List范围元素失败: %w", err)
	}
	return values, nil
}

// LLen 获取列表长度
func (c *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	cacheKey := c.keyBuilder.BuildKey(key)
	length, err := c.client.LlenCtx(ctx, cacheKey)
	if err != nil {
		return 0, fmt.Errorf("获取List长度失败: %w", err)
	}
	return int64(length), nil
}

// ==================== 私有辅助方法 ====================

// getFromCache 从缓存读取数据（内部方法）
//
// 返回值：剩余过期时间（TTL）和错误
func (c *RedisCache) getFromCache(ctx context.Context, cacheKey string, result interface{}) (time.Duration, error) {
	data, err := c.client.GetCtx(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, errcode.CacheNotFound
		}
		return 0, fmt.Errorf("读取缓存失败: %w", err)
	}

	// 检查是否是空值缓存
	if data == EmptyCacheValue {
		return 0, errcode.CacheEmpty
	}

	// 反序列化数据
	if err := json.Unmarshal([]byte(data), result); err != nil {
		return 0, fmt.Errorf("反序列化数据失败: %w", err)
	}

	// 获取剩余过期时间
	ttl, err := c.client.TtlCtx(ctx, cacheKey)
	if err != nil {
		return 0, nil // 忽略 TTL 获取失败
	}

	return time.Duration(ttl) * time.Second, nil
}

// unmarshalData 将数据反序列化到 result（内部方法）
func (c *RedisCache) unmarshalData(data interface{}, result interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("反序列化数据失败: %w", err)
	}

	return nil
}

// calculateExpiry 计算过期时间（基础时间 + 随机时间）（内部方法）
//
// 通过添加随机时间，防止缓存雪崩
func (c *RedisCache) calculateExpiry(opt *CacheOption) time.Duration {
	if opt.RandomExpiry == 0 {
		return opt.BaseExpiry
	}

	// 添加随机时间，防止缓存雪崩
	randomSeconds := rand.Int63n(int64(opt.RandomExpiry.Seconds()))
	return opt.BaseExpiry + time.Duration(randomSeconds)*time.Second
}

// tryAutoRefresh 尝试自动续期缓存（内部方法）
//
// 当缓存剩余时间低于阈值时，异步续期
func (c *RedisCache) tryAutoRefresh(ctx context.Context, cacheKey string, ttl time.Duration, opt *CacheOption) {
	// 计算总过期时间
	totalExpiry := opt.BaseExpiry + opt.RandomExpiry

	// 计算剩余时间百分比
	remainingPercent := float64(ttl) / float64(totalExpiry)

	// 如果剩余时间小于阈值，触发续期
	if remainingPercent < opt.AutoRefreshThreshold {
		// 异步续期，不阻塞主流程
		go func() {
			newExpiry := c.calculateExpiry(opt)
			_ = c.client.ExpireCtx(context.Background(), cacheKey, int(newExpiry.Seconds()))
		}()
	}
}
