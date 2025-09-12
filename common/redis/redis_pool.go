package redis_pool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// RedisPool 通用 Redis 连接池管理器
type RedisPool struct {
	client      *redis.Redis
	config      redis.RedisConf
	mu          sync.RWMutex
	initialized bool
}

var (
	globalRedisPool *RedisPool
	redisPoolOnce   sync.Once
)

// RedisPoolConfig Redis 连接池配置
type RedisPoolConfig struct {
	Host     string
	Type     string
	Pass     string
	Database int
}

// InitRedisPool 初始化 Redis 连接池（全局单例）
func InitRedisPool(config RedisPoolConfig) error {
	var err error
	redisPoolOnce.Do(func() {
		// 创建Redis客户端
		client := redis.New(config.Host, func(r *redis.Redis) {
			r.Type = config.Type
			r.Pass = config.Pass
		})

		// 测试连接
		if !client.Ping() {
			err = fmt.Errorf("Redis连接测试失败，请检查配置: Host=%s, Type=%s", config.Host, config.Type)
			return
		}

		globalRedisPool = &RedisPool{
			client: client,
			config: redis.RedisConf{
				Host: config.Host,
				Type: config.Type,
				Pass: config.Pass,
			},
			initialized: true,
		}
	})

	return err
}

// GetRedisClient 获取 Redis 客户端
func GetRedisClient() *redis.Redis {
	if globalRedisPool == nil {
		panic("Redis连接池未初始化，请先调用 InitRedisPool")
	}

	globalRedisPool.mu.RLock()
	defer globalRedisPool.mu.RUnlock()

	if !globalRedisPool.initialized {
		panic("Redis连接池初始化失败")
	}

	return globalRedisPool.client
}

// RedisKeyBuilder Redis Key 构建器
type RedisKeyBuilder struct {
	module   string
	business string
}

// NewRedisKeyBuilder 创建 Redis Key 构建器
func NewRedisKeyBuilder(module, business string) *RedisKeyBuilder {
	return &RedisKeyBuilder{
		module:   module,
		business: business,
	}
}

// BuildKey 构建 Redis Key
// 格式: [模块名]:[业务名]:[内容]
func (r *RedisKeyBuilder) BuildKey(content string) string {
	return fmt.Sprintf("%s:%s:%s", r.module, r.business, content)
}

// BuildKeys 批量构建 Redis Key
func (r *RedisKeyBuilder) BuildKeys(contents ...string) []string {
	keys := make([]string, len(contents))
	for i, content := range contents {
		keys[i] = r.BuildKey(content)
	}
	return keys
}

// ===== 工具函数 =====

// ParseRedisKey 解析Redis Key
func ParseRedisKey(key string) (module, business, content string, err error) {
	parts := strings.Split(key, ":")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("无效的Redis Key格式，期望格式: module:business:content")
	}

	module = parts[0]
	business = parts[1]
	content = strings.Join(parts[2:], ":") // 支持content中包含冒号

	return module, business, content, nil
}
