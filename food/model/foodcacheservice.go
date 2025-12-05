package model

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/hd2yao/ecshop/common/errcode"
	redisPool "github.com/hd2yao/ecshop/common/redis"
)

const (
	foodDetailCacheKeyFmt = "detail_%d"
	foodCreateLockKeyFmt  = "create_user_%d"
	foodUserListKeyFmt    = "user_%d_list"
	foodUserCountKeyFmt   = "count_%d"
	foodInfoHashKey       = "info_hash"
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
}

// NewFoodCacheService 创建美食缓存服务
func NewFoodCacheService(foodModel FoodModel) *FoodCacheService {
	return &FoodCacheService{
		foodModel: foodModel,
		cache:     redisPool.NewRedisCache("food", "info"),
		cacheOpt:  redisPool.DefaultCacheOption(),
	}
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

// toFoodDTO 将 Food 模型转换为 DTO
func toFoodDTO(food *Food) *FoodDTO {
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
	// 使用临时锁 key（基于 user_id，防止并发创建）
	lockKey := fmt.Sprintf(foodCreateLockKeyFmt, food.UserId)

	// 使用 UpdateWithMutex 保证并发安全
	var createdFood *Food
	var cacheKey string
	err := s.cache.UpdateWithMutex(ctx, lockKey, func() (interface{}, error) {
		// 1. 先插入数据库
		result, err := s.foodModel.Insert(ctx, food)
		if err != nil {
			return nil, fmt.Errorf("插入美食数据失败: %w", err)
		}

		// 2. 获取新插入的 food_id
		foodId, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("获取新美食ID失败: %w", err)
		}

		// 3. 查询新创建的美食信息（确保数据完整性）
		createdFood, err = s.foodModel.FindOne(ctx, foodId)
		if err != nil {
			return nil, fmt.Errorf("查询新创建的美食信息失败: %w", err)
		}

		// 4. 设置缓存key（用于后续写入缓存）
		cacheKey = fmt.Sprintf(foodDetailCacheKeyFmt, createdFood.Id)

		// 5. 返回数据用于更新缓存（注意：锁 key 和缓存 key 不同，需要在锁外写入）
		return nil, nil
	}, s.cacheOpt)

	if err != nil {
		return nil, err
	}

	// 在锁外写入缓存（因为锁key和缓存key不同）
	if cacheKey != "" {
		randomSeconds := rand.Int63n(int64(s.cacheOpt.RandomExpiry.Seconds()))
		expiry := s.cacheOpt.BaseExpiry + time.Duration(randomSeconds)*time.Second
		// 缓存写入失败不影响主流程
		_ = s.cache.Set(ctx, cacheKey, toFoodDTO(food), expiry)
	}

	return createdFood, nil
}

// updateFood 修改美食信息（带分布式锁和缓存）
func (s *FoodCacheService) updateFood(ctx context.Context, food *Food) (*Food, error) {
	// 使用美食ID作为锁key
	cacheKey := fmt.Sprintf(foodDetailCacheKeyFmt, food.Id)

	// 使用 UpdateWithMutex 保证并发安全和强一致性，并直接更新缓存
	var updatedFood *Food
	err := s.cache.UpdateWithMutex(ctx, cacheKey, func() (interface{}, error) {
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
		return toFoodDTO(updatedFood), nil
	}, s.cacheOpt)

	if err != nil {
		return nil, err
	}

	// 使相关列表缓存失效（详情缓存已在 UpdateWithMutex 中更新）
	s.invalidateRelatedCache(ctx, updatedFood.Id, updatedFood.UserId)

	return updatedFood, nil
}

// invalidateRelatedCache 使相关缓存失效
func (s *FoodCacheService) invalidateRelatedCache(ctx context.Context, foodId, userId int64) {
	// 删除用户美食列表缓存
	userListKey := fmt.Sprintf(foodUserListKeyFmt, userId)
	_ = s.cache.Delete(ctx, userListKey)

	// 删除用户美食总数缓存
	userCountKey := fmt.Sprintf(foodUserCountKeyFmt, userId)
	_ = s.cache.Delete(ctx, userCountKey)

	// 删除Hash缓存中的对应字段（如果使用Hash缓存）
	s.cache.HDel(ctx, foodInfoHashKey, fmt.Sprintf("%d", foodId))
}

// GetFoodById 根据 ID 获取美食信息（带缓存）
// 特性：
// 1. 优先从缓存读取，缓存命中后自动延期
// 2. 缓存未命中时使用分布式锁防止击穿
// 3. 数据不存在时设置空值缓存防止穿透
func (s *FoodCacheService) GetFoodById(ctx context.Context, foodId int64) (*FoodDTO, error) {
	cacheKey := fmt.Sprintf(foodDetailCacheKeyFmt, foodId)
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

		return toFoodDTO(food), nil
	}, s.cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &foodDTO, nil
}
