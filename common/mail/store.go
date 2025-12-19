package mail

import (
	"context"
	"time"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// MailCodeRedisStore 邮件验证码Redis存储
type MailCodeRedisStore struct {
	cache      *redisPool.RedisCache
	expiration time.Duration
}

// NewMailCodeRedisStore 创建邮件验证码Redis存储
func NewMailCodeRedisStore(module string, expiration time.Duration) *MailCodeRedisStore {
	if expiration == 0 {
		expiration = 10 * time.Minute // 默认10分钟，比图形验证码时间长一些
	}
	return &MailCodeRedisStore{
		cache:      redisPool.NewRedisCache(module, "mail"),
		expiration: expiration,
	}
}

// Set 存储邮件验证码
func (s *MailCodeRedisStore) Set(email, code string) error {
	ctx := context.Background()
	return s.cache.Set(ctx, email, code, s.expiration)
}

// Get 获取邮件验证码
func (s *MailCodeRedisStore) Get(email string, clear bool) string {
	ctx := context.Background()
	var val string
	if err := s.cache.Get(ctx, email, &val); err != nil {
		return ""
	}
	if clear {
		_ = s.cache.Delete(ctx, email)
	}
	return val
}

// Verify 验证邮件验证码
func (s *MailCodeRedisStore) Verify(email, code string, clear bool) bool {
	v := s.Get(email, clear)
	return v == code
}

// Exists 检查邮件验证码是否存在
func (s *MailCodeRedisStore) Exists(email string) bool {
	ctx := context.Background()
	exists, _ := s.cache.Exists(ctx, email)
	return exists
}

// GetTTL 获取验证码剩余有效时间
func (s *MailCodeRedisStore) GetTTL(email string) time.Duration {
	ctx := context.Background()
	ttl, err := s.cache.GetExpire(ctx, email)
	if err != nil {
		return 0
	}
	if ttl <= 0 {
		return 0
	}
	return time.Duration(ttl) * time.Second
}
