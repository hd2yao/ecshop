package captcha

import (
	"context"
	"time"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// CaptchaRedisStore 实现base64Captcha.Store接口的Redis存储
type CaptchaRedisStore struct {
	cache      *redisPool.RedisCache
	expiration time.Duration
}

// NewCaptchaRedisStore 创建Redis存储
func NewCaptchaRedisStore(module string, expiration time.Duration) *CaptchaRedisStore {
	if expiration == 0 {
		expiration = 5 * time.Minute // 默认5分钟
	}
	return &CaptchaRedisStore{
		cache:      redisPool.NewRedisCache(module, "captcha"),
		expiration: expiration,
	}
}

// Set 存储验证码
func (s *CaptchaRedisStore) Set(id, value string) error {
	ctx := context.Background()
	return s.cache.Set(ctx, id, value, s.expiration)
}

// Get 获取验证码
func (s *CaptchaRedisStore) Get(id string, clear bool) string {
	ctx := context.Background()
	var val string
	if err := s.cache.Get(ctx, id, &val); err != nil {
		return ""
	}
	if clear {
		_ = s.cache.Delete(ctx, id)
	}
	return val
}

// Verify 验证验证码
func (s *CaptchaRedisStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == answer
}
