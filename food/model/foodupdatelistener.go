package model

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/common/rocketmq"
)

const pageSize = 20

// FoodUpdateListener 美食更新监听器
type FoodUpdateListener struct {
	foodModel FoodModel
	logger    logx.Logger
	cache     *redisPool.RedisCache
	lock      *redisPool.LockExecutor
}

// NewFoodUpdateListener 创建美食更新监听器
func NewFoodUpdateListener(foodModel FoodModel, cache *redisPool.RedisCache) *FoodUpdateListener {
	return &FoodUpdateListener{
		foodModel: foodModel,
		logger:    logx.WithContext(context.Background()),
		cache:     cache,
		lock:      redisPool.NewLockExecutor(),
	}
}

// ConsumeMessage 消费消息
func (l *FoodUpdateListener) ConsumeMessage(ctx context.Context, msgs []*rocketmq.Message) (rocketmq.ConsumeStatus, error) {
	l.logger.Infof("美食 MQ 消费消息-收到 %d 条消息", len(msgs))

	for i, msg := range msgs {
		l.logger.Infof("美食 MQ 消费消息-第 %d/%d 条: MessageID=%s, Topic=%s, Tag=%s, Key=%s",
			i+1, len(msgs), msg.MessageID, msg.Topic, msg.Tag, msg.Key)

		// 解析消息
		var foodMsg FoodUpdateMQMessage
		if err := msg.UnmarshalBody(&foodMsg); err != nil {
			l.logger.Errorf("美食 MQ 消费消息-解析消息体失败: MessageID=%s, error=%v, body=%s",
				msg.MessageID, err, msg.GetBodyString())
			continue
		}

		userID := foodMsg.UserID
		foodID := foodMsg.FoodID

		l.logger.Infof("美食 MQ 消费消息-开始处理: foodID=%d, userID=%d", foodID, userID)

		// 使用分布式锁，同一个用户同时间只能操作一次，避免重复请求
		lockKey := fmt.Sprintf(FoodUpdateLockKeyFmt, userID)

		// 通过阻塞的方式进行加锁，跟新增/更新美食操作互斥，并发操作时进行阻塞
		err := l.lock.ExecuteWithLock(ctx, lockKey, 10*time.Second, 3*time.Second, func() error {
			return l.rebuildUserPageCache(ctx, userID)
		})

		if err != nil {
			l.logger.Errorf("美食 MQ 消费消息-处理失败: foodID=%d, userID=%d, error=%v", foodID, userID, err)
			return rocketmq.ConsumeLater, err
		}

		l.logger.Infof("美食 MQ 消费消息-处理成功: foodID=%d, userID=%d", foodID, userID)
	}

	l.logger.Infof("美食 MQ 消费消息-批量处理完成: 共 %d 条消息", len(msgs))
	return rocketmq.ConsumeSuccess, nil
}

// rebuildUserPageCache 重建用户分页缓存
func (l *FoodUpdateListener) rebuildUserPageCache(ctx context.Context, userID int64) error {
	// 1. 计算我的美食列表总数
	totalKey := fmt.Sprintf(FoodMyTotalKeyFmt, userID)
	var total int64
	if err := l.cache.Get(ctx, totalKey, &total); err != nil {
		// 如果缓存中没有总数，从数据库查询
		_, total, err = l.foodModel.FindListByUserId(ctx, userID, 1, pageSize)
		if err != nil {
			return fmt.Errorf("查询美食总数失败: %w", err)
		}
		// 写入总数缓存
		cacheOpt := redisPool.DefaultCacheOption()
		_ = l.cache.Set(ctx, totalKey, total, cacheOpt.BaseExpiry)
	}

	l.logger.Infof("美食 MQ 异步更新-key:%s, value:%d", totalKey, total)

	// 计算总分页数
	pageNums := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pageNums == 0 {
		pageNums = 1
	}

	// 2. 循环构建美食分页缓存（只重建已存在的缓存页）
	for page := 1; page <= pageNums; page++ {
		// 美食分页缓存 key
		pageKey := fmt.Sprintf(FoodMyPageKeyFmt, userID, page)

		// 先从缓存中获取一下
		var existingCache MyFoodPageCache
		if err := l.cache.Get(ctx, pageKey, &existingCache); err != nil {
			// 如果分页缓存为空，跳过（只重建已存在的缓存）
			continue
		}

		// 只有缓存中存在这一页的数据才会去数据库读取最新的美食信息去更新
		// 查询我的未删除美食列表
		foods, _, err := l.foodModel.FindListByUserId(ctx, userID, int32(page), pageSize)
		if err != nil {
			l.logger.Errorf("美食 MQ 异步更新-从数据库获取美食列表失败: page=%d, error=%v", page, err)
			continue
		}

		// 转换为 DTO
		result := MyFoodPageCache{
			List:  make([]FoodDTO, 0, len(foods)),
			Total: total,
		}

		for _, f := range foods {
			if f == nil {
				continue
			}
			if dto := ToFoodDTO(f); dto != nil {
				result.List = append(result.List, *dto)
			}
		}

		// 写入分页缓存数据
		cacheOpt := redisPool.DefaultCacheOption()
		expiry := cacheOpt.BaseExpiry + time.Duration(rand.Int63n(int64(cacheOpt.RandomExpiry.Seconds())))*time.Second
		if err := l.cache.Set(ctx, pageKey, result, expiry); err != nil {
			l.logger.Errorf("美食 MQ 异步更新-写入分页缓存失败: page=%d, error=%v", page, err)
			continue
		}

		l.logger.Infof("美食 MQ 异步更新-重建分页缓存成功: userID=%d, page=%d, size=%d", userID, page, len(result.List))
	}

	return nil
}

