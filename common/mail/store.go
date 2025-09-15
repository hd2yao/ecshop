package mail

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"

	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// MailCodeRedisStore 邮件验证码Redis存储
type MailCodeRedisStore struct {
	redis      *redis.Redis
	keyBuilder *redisPool.RedisKeyBuilder
	expiration time.Duration
}

// NewMailCodeRedisStore 创建邮件验证码Redis存储
func NewMailCodeRedisStore(module string, expiration time.Duration) *MailCodeRedisStore {
	if expiration == 0 {
		expiration = 10 * time.Minute // 默认10分钟，比图形验证码时间长一些
	}
	return &MailCodeRedisStore{
		redis:      redisPool.GetRedisClient(),
		keyBuilder: redisPool.NewRedisKeyBuilder(module, "mail"),
		expiration: expiration,
	}
}

// Set 存储邮件验证码
func (s *MailCodeRedisStore) Set(email, code string) error {
	key := s.keyBuilder.BuildKey(email)
	return s.redis.Setex(key, code, int(s.expiration.Seconds()))
}

// Get 获取邮件验证码
func (s *MailCodeRedisStore) Get(email string, clear bool) string {
	key := s.keyBuilder.BuildKey(email)
	val, err := s.redis.Get(key)
	if err != nil {
		return ""
	}
	if clear {
		s.redis.Del(key)
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
	key := s.keyBuilder.BuildKey(email)
	exists, _ := s.redis.Exists(key)
	return exists
}

// GetTTL 获取验证码剩余有效时间
func (s *MailCodeRedisStore) GetTTL(email string) time.Duration {
	key := s.keyBuilder.BuildKey(email)
	ttl, err := s.redis.Ttl(key)
	if err != nil {
		return 0
	}
	return time.Duration(ttl) * time.Second
}
