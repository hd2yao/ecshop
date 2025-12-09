package redis_pool

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/hd2yao/ecshop/common/errcode"
)

// ==================== 类型定义 ====================

// RedisLock 分布式锁
//
// 特性：
//   - UUID 防误删：使用 UUID 作为锁值，防止误删其他客户端的锁
//   - 看门狗续期：支持自动续期，适用于长时间任务
//   - 原子性保证：使用 Lua 脚本保证加锁/解锁/续期的原子性
type RedisLock struct {
	client    *redis.Redis
	key       string
	value     string // 使用 UUID 作为锁的唯一标识，防止误删
	expiry    time.Duration
	watchDog  bool          // 是否启用看门狗自动续期
	stopWatch chan struct{} // 用于停止看门狗的通道
}

// RedisLockOption 分布式锁配置选项
type RedisLockOption func(*RedisLock)

// LockExecutor 分布式锁执行器
type LockExecutor struct {
	client *redis.Redis
}

// ==================== 构造函数 ====================

// NewRedisLock 创建分布式锁实例
//
// 参数：
//   - key: 锁的键名
//   - expiry: 锁的过期时间
//   - opts: 可选配置（如 WithWatchDog()）
//
// 示例：
//
//	lock := NewRedisLock("order:123", 10*time.Second)
//	lock := NewRedisLock("payment:456", 30*time.Second, WithWatchDog())
func NewRedisLock(key string, expiry time.Duration, opts ...RedisLockOption) *RedisLock {
	lock := &RedisLock{
		client:    GetRedisClient(),
		key:       key,
		value:     uuid.New().String(),
		expiry:    expiry,
		watchDog:  false,
		stopWatch: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(lock)
	}

	return lock
}

// WithWatchDog 启用看门狗自动续期（适用于长时间任务）
//
// 使用场景：支付流程、批量处理等执行时间不确定的任务
//
// 示例：
//
//	lock := NewRedisLock("payment:123", 30*time.Second, WithWatchDog())
func WithWatchDog() RedisLockOption {
	return func(l *RedisLock) {
		l.watchDog = true
	}
}

// NewLockExecutor 创建分布式锁执行器
func NewLockExecutor() *LockExecutor {
	return &LockExecutor{
		client: GetRedisClient(),
	}
}

// ==================== 锁执行器方法 ====================

// ExecuteWithLock 在分布式锁保护下执行函数（阻塞式）
//
// 参数：
//   - lockKey: 锁的键名
//   - expiry: 锁的过期时间
//   - timeout: 获取锁的超时时间（0 表示无限等待）
//   - fn: 要执行的业务逻辑函数
//   - opts: 可选配置（如 WithWatchDog()）
//
// 特点：
//   - 自动处理加锁和解锁
//   - 使用 defer 确保锁一定会被释放
//   - 适合需要阻塞等待获取锁的场景
//
// 使用场景：强一致性操作、必须获取锁才能执行的任务
//
// 示例：
//
//	executor := NewLockExecutor()
//	err := executor.ExecuteWithLock(ctx, "order:123", 10*time.Second, 3*time.Second, func() error {
//	    // 处理订单
//	    return processOrder()
//	})
func (e *LockExecutor) ExecuteWithLock(
	ctx context.Context,
	lockKey string,
	expiry time.Duration,
	timeout time.Duration,
	fn func() error,
	opts ...RedisLockOption,
) error {
	lock := NewRedisLock(lockKey, expiry, opts...)

	// 获取锁
	if err := lock.Lock(ctx, timeout); err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}

	// 确保释放锁
	defer func() {
		if err := lock.Unlock(ctx); err != nil {
			// 这里可以记录日志
		}
	}()

	// 执行业务逻辑
	return fn()
}

func (e *LockExecutor) ExecuteWithTryLock(
	ctx context.Context,
	lockKey string,
	expiry time.Duration,
	fn func() error,
	opts ...RedisLockOption,
) (executed bool, err error) {
	lock := NewRedisLock(lockKey, expiry, opts...)

	// 尝试获取锁
	ok, err := lock.TryLock(ctx)
	if err != nil {
		return false, fmt.Errorf("尝试获取锁失败: %w", err)
	}

	if !ok {
		return false, nil
	}

	// 确保释放锁
	defer func() {
		if unlockErr := lock.Unlock(ctx); unlockErr != nil {
			// 这里可以记录日志
		}
	}()

	// 执行业务逻辑
	return true, fn()
}

// ==================== 基础加锁方法 ====================

// TryLock 尝试获取锁（非阻塞）
//
// 返回值：
//   - true: 获取锁成功
//   - false: 锁已被其他客户端持有
//
// 使用场景：防止重复操作、快速失败等场景
//
// 示例：
//
//	ok, err := lock.TryLock(ctx)
//	if !ok {
//	    return errors.New("请勿重复操作")
//	}
func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	// 使用 SET NX EX 命令实现原子性加锁
	ok, err := l.client.SetnxExCtx(ctx, l.key, l.value, int(l.expiry.Seconds()))
	if err != nil {
		return false, fmt.Errorf("尝试获取锁失败: %w", err)
	}

	// 如果获取锁成功且启用了看门狗，启动自动续期
	if ok && l.watchDog {
		l.startWatchDog(ctx)
	}

	return ok, nil
}

// TryLockWithTimeout 尝试获取锁（带超时，阻塞式）
//
// 参数：
//   - timeout: 等待获取锁的超时时间
//
// 返回值：
//   - true: 获取锁成功
//   - false: 超时或锁已被其他客户端持有（不返回错误）
//
// 使用场景：需要等待但不想无限等待的场景
//
// 示例：
//
//	ok, err := lock.TryLockWithTimeout(ctx, 3*time.Second)
//	if !ok {
//	    return errors.New("获取锁超时")
//	}
func (l *RedisLock) TryLockWithTimeout(ctx context.Context, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond) // 每 50ms 重试一次
	defer ticker.Stop()

	for {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}

		// 检查超时
		if time.Now().After(deadline) {
			return false, nil // 超时返回 false，不返回错误
		}

		// 等待下次重试
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

// Lock 阻塞式获取锁（带超时）
//
// 参数：
//   - timeout: 获取锁的超时时间
//   - timeout > 0: 等待指定时间后超时，返回错误
//   - timeout = 0: 无限等待（直到获取到锁或 context 取消）
//
// 注意：
//   - 如果 timeout = 0，会自动启用看门狗，防止死锁
//   - 如果 timeout > 0，建议根据任务执行时间设置合理的过期时间
//
// 使用场景：
//   - timeout = 0: 必须获取到锁的关键任务
//   - timeout > 0: 有超时限制的任务, 超时会报错
//
// 示例：
//
//	// 无限等待（自动启用看门狗）
//	err := lock.Lock(ctx, 0)
//
//	// 有限等待（最多等待 5 秒）
//	err := lock.Lock(ctx, 5*time.Second)
func (l *RedisLock) Lock(ctx context.Context, timeout time.Duration) error {
	// 如果 timeout = 0，无限等待
	if timeout == 0 {
		// 无限等待模式，必须启用看门狗
		if !l.watchDog {
			// 自动启用看门狗，防止死锁
			l.watchDog = true
		}

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			ok, err := l.TryLock(ctx)
			if err != nil {
				return err
			}

			if ok {
				return nil
			}

			// 等待下次重试
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				continue
			}
		}
	}

	// timeout > 0，有限等待
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		ok, err := l.TryLock(ctx)
		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		// 检查超时
		if time.Now().After(deadline) {
			return errcode.LockFailed
		}

		// 等待下次重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

// LockWithExpiry 获取锁并设置过期时间（不启用看门狗）
//
// 参数：
//   - expiry: 锁的过期时间，到期后自动释放
//
// 注意：
//   - 此方法不启用看门狗，锁会在指定时间后自动过期
//   - 适用于明确知道任务执行时间的场景
//
// 使用场景：订单处理、定时任务等有明确执行时间的场景
//
// 示例：
//
//	// 处理订单，最多需要 30 秒
//	err := lock.LockWithExpiry(ctx, 30*time.Second)
//	defer lock.Unlock(ctx)
func (l *RedisLock) LockWithExpiry(ctx context.Context, expiry time.Duration) error {
	// 临时保存原始过期时间
	originalExpiry := l.expiry
	l.expiry = expiry

	// 确保不启用看门狗
	originalWatchDog := l.watchDog
	l.watchDog = false

	// 尝试获取锁
	ok, err := l.TryLock(ctx)
	if err != nil {
		// 恢复原始设置
		l.expiry = originalExpiry
		l.watchDog = originalWatchDog
		return err
	}

	if !ok {
		// 恢复原始设置
		l.expiry = originalExpiry
		l.watchDog = originalWatchDog
		return errcode.LockFailed
	}

	// 恢复原始设置（不影响已获取的锁）
	l.expiry = originalExpiry
	l.watchDog = originalWatchDog

	return nil
}

// ==================== 解锁方法 ====================

// Unlock 释放锁
//
// 只有持有锁的客户端才能释放锁（通过 UUID 验证）
//
// 注意：
//   - 使用 Lua 脚本保证原子性
//   - 会自动停止看门狗
//   - 防止重复关闭 channel
//
// 示例：
//
//	err := lock.Unlock(ctx)
//	if err != nil {
//	    log.Errorf("释放锁失败: %v", err)
//	}
func (l *RedisLock) Unlock(ctx context.Context) error {
	// 停止看门狗
	if l.watchDog {
		select {
		case <-l.stopWatch:
			// 已经关闭，不需要再次关闭
		default:
			close(l.stopWatch)
		}
	}

	// 使用 Lua 脚本保证原子性：只有持有锁的客户端才能释放锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.EvalCtx(ctx, script, []string{l.key}, l.value)
	if err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}

	if result == int64(0) {
		return errcode.LockNotHeld
	}

	return nil
}

// ==================== 锁状态检查方法 ====================

// IsLocked 检查锁是否被持有（不区分持有者）
//
// 返回值：
//   - true: 锁已被持有（可能是自己或其他客户端）
//   - false: 锁未被持有
//
// 使用场景：检查任务是否正在执行、避免重复操作等
//
// 示例：
//
//	isLocked, err := lock.IsLocked(ctx)
//	if isLocked {
//	    log.Info("任务正在执行中，跳过")
//	    return
//	}
func (l *RedisLock) IsLocked(ctx context.Context) (bool, error) {
	exists, err := l.client.ExistsCtx(ctx, l.key)
	if err != nil {
		return false, fmt.Errorf("检查锁状态失败: %w", err)
	}
	return exists, nil
}

// IsHeldByCurrentThread 检查锁是否被当前实例持有
//
// 返回值：
//   - true: 锁被当前实例持有
//   - false: 锁未被持有或被其他客户端持有
//
// 使用场景：可重入锁检查、安全操作验证等
//
// 示例：
//
//	isHeld, err := lock.IsHeldByCurrentThread(ctx)
//	if !isHeld {
//	    return errors.New("锁未被当前实例持有，无法操作")
//	}
func (l *RedisLock) IsHeldByCurrentThread(ctx context.Context) (bool, error) {
	value, err := l.client.GetCtx(ctx, l.key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // 锁不存在
		}
		return false, fmt.Errorf("检查锁持有者失败: %w", err)
	}

	// 检查锁的值是否匹配当前实例的 UUID
	return value == l.value, nil
}

// ==================== 锁管理方法 ====================

// Refresh 刷新锁的过期时间
//
// 只有持有锁的客户端才能续期（通过 UUID 验证）
//
// 使用场景：手动续期、看门狗续期等
//
// 注意：通常不需要手动调用，看门狗会自动处理
func (l *RedisLock) Refresh(ctx context.Context) error {
	// 使用 Lua 脚本保证原子性：只有持有锁的客户端才能续期
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.EvalCtx(ctx, script, []string{l.key}, l.value, int(l.expiry.Seconds()))
	if err != nil {
		return fmt.Errorf("刷新锁失败: %w", err)
	}

	if result == int64(0) {
		return errcode.LockExpired
	}

	return nil
}

// ==================== 私有辅助方法 ====================

// startWatchDog 启动看门狗自动续期（内部方法）
//
// 工作原理：
//   - 在锁过期时间的 1/3 处进行续期
//   - 如果续期失败（锁已过期或被删除），自动停止
//   - 通过 stopWatch channel 控制停止
func (l *RedisLock) startWatchDog(ctx context.Context) {
	go func() {
		// 在锁过期时间的 1/3 处进行续期
		ticker := time.NewTicker(l.expiry / 3)
		defer ticker.Stop()

		for {
			select {
			case <-l.stopWatch:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := l.Refresh(ctx); err != nil {
					// 续期失败，停止看门狗
					return
				}
			}
		}
	}()
}
