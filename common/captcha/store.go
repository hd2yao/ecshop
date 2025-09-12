package captcha

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// CaptchaRedisStore 实现base64Captcha.Store接口的Redis存储
type CaptchaRedisStore struct {
	redis      *redis.Redis
	keyBuilder *redisPool.RedisKeyBuilder
	expiration time.Duration
}

// NewCaptchaRedisStore 创建Redis存储
func NewCaptchaRedisStore(module string, expiration time.Duration) *CaptchaRedisStore {
	if expiration == 0 {
		expiration = 5 * time.Minute // 默认5分钟
	}
	return &CaptchaRedisStore{
		redis:      redisPool.GetRedisClient(),
		keyBuilder: redisPool.NewRedisKeyBuilder(module, "captcha"),
		expiration: expiration,
	}
}

// Set 存储验证码
func (s *CaptchaRedisStore) Set(id, value string) error {
	key := s.keyBuilder.BuildKey(id)
	return s.redis.Setex(key, value, int(s.expiration.Seconds()))
}

// Get 获取验证码
func (s *CaptchaRedisStore) Get(id string, clear bool) string {
	key := s.keyBuilder.BuildKey(id)
	val, err := s.redis.Get(key)
	if err != nil {
		return ""
	}
	if clear {
		s.redis.Del(key)
	}
	return val
}

// Verify 验证验证码
func (s *CaptchaRedisStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == answer
}
