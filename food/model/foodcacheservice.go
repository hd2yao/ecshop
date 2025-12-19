package model

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/common/rocketmq"
)

const (
	FoodDetailCacheKeyFmt = "detail_%d"
	FoodCreateLockKeyFmt  = "create_lock_%d"
	FoodUpdateLockKeyFmt  = "update_user_%d"

	// 我的美食分页缓存
	FoodMyPageKeyFmt  = "user_page_%d_%d" // userId, page
	FoodMyTotalKeyFmt = "user_total_%d"   // userId（自增统计）
)

// FoodCacheService 美食缓存服务
// 封装了美食数据的缓存逻辑，包括：
// 1. 热点数据自动延期
// 2. 缓存穿透防护
// 3. 缓存击穿防护（分布式锁）
// 4. 数据库和缓存双写一致性
// 5. 高并发场景下的锁优化
type FoodCacheService struct {
	foodModel FoodModel
	cache     *redisPool.RedisCache
	cacheOpt  *redisPool.CacheOption
	lock      *redisPool.LockExecutor
}

// NewFoodCacheService 创建美食缓存服务
// 如果 cache 为 nil，则创建新的 RedisCache 实例
func NewFoodCacheService(foodModel FoodModel, cache *redisPool.RedisCache) *FoodCacheService {
	if cache == nil {
		cache = redisPool.NewRedisCache("food", "info")
	}
	return &FoodCacheService{
		foodModel: foodModel,
		cache:     cache,
		cacheOpt:  redisPool.DefaultCacheOption(),
		lock:      redisPool.NewLockExecutor(),
	}
}

// MyFoodPageCache 我的美食列表分页缓存结构
type MyFoodPageCache struct {
	List  []FoodDTO `json:"list"`
	Total int64     `json:"total"`
}

// FoodDTO 美食数据传输对象
type FoodDTO struct {
	Id             int64     `json:"id"`
	UserId         int64     `json:"user_id"`
	FoodName       string    `json:"food_name"`
	FoodDes        string    `json:"food_des"`
	FoodCateTag    int64     `json:"food_cate_tag"`
	FoodUrl        string    `json:"food_url"`
	FoodVideoUrl   string    `json:"food_video_url"`
	FoodTime       int64     `json:"food_time"`
	FoodDifficulty int64     `json:"food_difficulty"`
	FoodDetail     string    `json:"food_detail"` // JSON 字符串
	FoodList       string    `json:"food_list"`   // JSON 字符串
	FoodStatus     int64     `json:"food_status"`
	FoodSkuIds     string    `json:"food_sku_ids"` // JSON 字符串（可为空）
	FoodCreatetime time.Time `json:"food_createtime"`
	FoodUpdatetime time.Time `json:"food_updatetime"`
}

// ToFoodDTO 将 Food 模型转换为 DTO
func ToFoodDTO(food *Food) *FoodDTO {
	if food == nil {
		return nil
	}

	dto := &FoodDTO{
		Id:             food.Id,
		UserId:         food.UserId,
		FoodName:       food.FoodName,
		FoodDes:        food.FoodDes,
		FoodCateTag:    food.FoodCateTag,
		FoodUrl:        food.FoodUrl,
		FoodVideoUrl:   food.FoodVideoUrl,
		FoodTime:       food.FoodTime,
		FoodDifficulty: food.FoodDifficulty,
		FoodDetail:     food.FoodDetail,
		FoodList:       food.FoodList,
		FoodStatus:     food.FoodStatus,
		FoodCreatetime: food.FoodCreatetime,
		FoodUpdatetime: food.FoodUpdatetime,
	}

	// 处理 FoodSkuIds（sql.NullString）
	if food.FoodSkuIds.Valid {
		dto.FoodSkuIds = food.FoodSkuIds.String
	}

	return dto
}

// CreateOrUpdateFood 新增/修改美食信息（带分布式锁和缓存）
// 使用分布式互斥锁防止重复灌入数据，写完数据库+Redis后释放锁,过期时间为2天+随机数
func (s *FoodCacheService) CreateOrUpdateFood(ctx context.Context, food *Food) (*Food, error) {
	// 判断是新增还是修改
	if food.Id > 0 {
		// 修改美食信息
		return s.updateFood(ctx, food)
	} else {
		// 新增美食信息
		return s.createFood(ctx, food)
	}
}

// createFood 新增美食信息（带分布式锁和缓存）
func (s *FoodCacheService) createFood(ctx context.Context, food *Food) (*Food, error) {
	// 使用用户级分布式锁，统一串行写/读重建
	lockKey := fmt.Sprintf(FoodUpdateLockKeyFmt, food.UserId)

	var createdFood *Food

	err := s.lock.ExecuteWithLock(ctx, lockKey, 10*time.Second, 3*time.Second, func() error {
		// 1. 先插入数据库
		result, err := s.foodModel.Insert(ctx, food)
		if err != nil {
			return fmt.Errorf("插入美食数据失败: %w", err)
		}

		// 2. 获取新插入的 food_id
		foodId, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("获取新美食ID失败: %w", err)
		}

		// 3. 查询新创建的美食信息（确保数据完整性）
		createdFood, err = s.foodModel.FindOne(ctx, foodId)
		if err != nil {
			return fmt.Errorf("查询新创建的美食信息失败: %w", err)
		}

		// 4. 写入明细缓存（锁内写，保证与写库一致）
		cacheKey := fmt.Sprintf(FoodDetailCacheKeyFmt, createdFood.Id)
		randomSeconds := rand.Int63n(int64(s.cacheOpt.RandomExpiry.Seconds()))
		expiry := s.cacheOpt.BaseExpiry + time.Duration(randomSeconds)*time.Second
		_ = s.cache.Set(ctx, cacheKey, ToFoodDTO(createdFood), expiry)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 维护总数自增（新增时）
	s.incrMyTotal(ctx, createdFood.UserId, 1)

	// 发送 RocketMQ 消息，异步更新分页缓存
	s.sendFoodUpdateMessage(ctx, createdFood.Id, createdFood.UserId)

	return createdFood, nil
}

// updateFood 修改美食信息（带分布式锁和缓存）
func (s *FoodCacheService) updateFood(ctx context.Context, food *Food) (*Food, error) {
	// 使用用户级分布式锁，统一串行写/读重建
	lockKey := fmt.Sprintf(FoodUpdateLockKeyFmt, food.UserId)
	cacheKey := fmt.Sprintf(FoodDetailCacheKeyFmt, food.Id)

	var updatedFood *Food
	err := s.lock.ExecuteWithLock(ctx, lockKey, 10*time.Second, 3*time.Second, func() error {
		// 使用 UpdateWithMutex 保证并发安全和强一致性，并直接更新缓存
		return s.cache.UpdateWithMutex(ctx, cacheKey, func() (interface{}, error) {
			// 1. 先更新数据库
			if err := s.foodModel.Update(ctx, food); err != nil {
				return nil, fmt.Errorf("更新美食数据失败: %w", err)
			}

			// 2. 查询更新后的美食信息（确保数据完整性）
			var err error
			updatedFood, err = s.foodModel.FindOne(ctx, food.Id)
			if err != nil {
				return nil, fmt.Errorf("查询更新后的美食信息失败: %w", err)
			}

			// 3. 返回更新后的数据，用于直接更新缓存（而不是删除）
			return ToFoodDTO(updatedFood), nil
		}, s.cacheOpt)
	})

	if err != nil {
		return nil, err
	}

	// 发送 RocketMQ 消息，异步更新分页缓存
	s.sendFoodUpdateMessage(ctx, updatedFood.Id, updatedFood.UserId)

	return updatedFood, nil
}

// GetFoodById 根据 ID 获取美食信息（带缓存）
// 特性：
// 1. 优先从缓存读取，缓存命中后自动延期
// 2. 缓存未命中时使用分布式锁防止击穿
// 3. 数据不存在时设置空值缓存防止穿透
func (s *FoodCacheService) GetFoodById(ctx context.Context, foodId int64) (*FoodDTO, error) {
	cacheKey := fmt.Sprintf(FoodDetailCacheKeyFmt, foodId)
	var foodDTO FoodDTO

	err := s.cache.GetWithLoader(ctx, cacheKey, &foodDTO, func() (interface{}, error) {
		// 从数据库加载美食数据
		food, err := s.foodModel.FindOne(ctx, foodId)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, nil // 返回 nil 表示数据不存在
			}
			return nil, err
		}

		return ToFoodDTO(food), nil
	}, s.cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &foodDTO, nil
}

// GetMyFoodPage 按需构建并获取“我的美食”分页缓存
func (s *FoodCacheService) GetMyFoodPage(ctx context.Context, userId int64, page, pageSize int32) ([]FoodDTO, int64, error) {
	pageKey := fmt.Sprintf(FoodMyPageKeyFmt, userId, page)
	totalKey := fmt.Sprintf(FoodMyTotalKeyFmt, userId)
	var pageCache MyFoodPageCache
	lockKey := fmt.Sprintf(FoodUpdateLockKeyFmt, userId)

	err := s.lock.ExecuteWithLock(ctx, lockKey, 5*time.Second, 2*time.Second, func() error {
		// 使用 GetWithLoader 按需构建缓存，内置防击穿 + 自动续期
		return s.cache.GetWithLoader(ctx, pageKey, &pageCache, func() (interface{}, error) {
			foods, total, err := s.foodModel.FindListByUserId(ctx, userId, page, pageSize)
			if err != nil {
				return nil, err
			}

			result := MyFoodPageCache{
				List:  make([]FoodDTO, 0, len(foods)),
				Total: total, // 默认使用数据库返回的总数（首次）
			}

			for _, f := range foods {
				if f == nil {
					continue
				}
				if dto := ToFoodDTO(f); dto != nil {
					result.List = append(result.List, *dto)
				}
			}

			// 如果已有总数自增缓存，则覆盖为自增值；否则写入自增总数
			var cachedTotal int64
			if err := s.cache.Get(ctx, totalKey, &cachedTotal); err == nil {
				result.Total = cachedTotal
			} else {
				// 初始化总数缓存，过期时间与分页一致
				_ = s.cache.Set(ctx, totalKey, total, s.cacheOpt.BaseExpiry)
			}

			return result, nil
		}, s.cacheOpt)
	})

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return []FoodDTO{}, 0, nil
		}
		return nil, 0, err
	}

	return pageCache.List, pageCache.Total, nil
}

// incrMyTotal 维护我的美食总数（自增/自减）
func (s *FoodCacheService) incrMyTotal(ctx context.Context, userId int64, delta int64) {
	key := fmt.Sprintf(FoodMyTotalKeyFmt, userId)
	if delta == 0 {
		return
	}
	if delta > 0 {
		_, _ = s.cache.Increment(ctx, key, delta)
		return
	}
	_, _ = s.cache.Decrement(ctx, key, -delta)
}

// sendFoodUpdateMessage 发送美食更新消息到 RocketMQ
func (s *FoodCacheService) sendFoodUpdateMessage(ctx context.Context, foodID, userID int64) {
	// 使用 goroutine 异步发送，避免阻塞主流程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// RocketMQ 未初始化时，忽略错误
			}
		}()

		// 获取 RocketMQ 实例（可能未初始化，使用 recover 捕获 panic）
		rocketMQ := rocketmq.GetRocketMQ()

		// 创建消息
		msg := FoodUpdateMQMessage{
			FoodID: foodID,
			UserID: userID,
		}

		// 发送消息（异步发送，失败不影响主流程）
		logx.Infof("美食 MQ 生产消息-准备发送: Topic=food_cache_update, Tag=update, foodID=%d, userID=%d", foodID, userID)
		result, err := rocketMQ.SendMessage(context.Background(), "food_cache_update", "update", msg)
		if err != nil {
			// 记录日志，但不影响主流程
			logx.Errorf("美食 MQ 生产消息-发送失败: foodID=%d, userID=%d, error=%v", foodID, userID, err)
		} else {
			logx.Infof("美食 MQ 生产消息-发送成功: foodID=%d, userID=%d, MessageID=%s, QueueOffset=%d",
				foodID, userID, result.MessageID, result.QueueOffset)
		}
	}()
}
