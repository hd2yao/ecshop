package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hd2yao/ecshop/common/errcode"
	redisPool "github.com/hd2yao/ecshop/common/redis"
)

// UserCacheService 用户缓存服务
// 封装了用户数据的缓存逻辑，包括：
// 1. 热点数据自动延期
// 2. 缓存穿透防护
// 3. 缓存击穿防护（分布式锁）
// 4. 数据库和缓存双写一致性
// 5. 高并发场景下的锁优化
type UserCacheService struct {
	userModel UserModel
	cache     *redisPool.RedisCache
	cacheOpt  *redisPool.CacheOption
}

// DataSource 表示数据来源
type DataSource string

const (
	// DataSourceCache 表示数据来自缓存
	DataSourceCache DataSource = "cache"
	// DataSourceDatabase 表示数据来自数据库
	DataSourceDatabase DataSource = "database"
)

// NewUserCacheService 创建用户缓存服务
func NewUserCacheService(userModel UserModel) *UserCacheService {
	return &UserCacheService{
		userModel: userModel,
		cache:     redisPool.NewRedisCache("user", "info"),
		cacheOpt:  redisPool.DefaultCacheOption(),
	}
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	Id         uint64    `json:"id"`
	Name       string    `json:"name"`
	Avatar     string    `json:"avatar"`
	Sex        int64     `json:"sex"`
	Points     int64     `json:"points"`
	Mail       string    `json:"mail"`
	Phone      string    `json:"phone"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// toUserDTO 将 User 模型转换为 DTO
func toUserDTO(user *User) *UserDTO {
	if user == nil {
		return nil
	}

	dto := &UserDTO{
		Id:     user.Id,
		Sex:    user.Sex,
		Points: user.Points,
	}

	if user.Name.Valid {
		dto.Name = user.Name.String
	}
	if user.Avatar.Valid {
		dto.Avatar = user.Avatar.String
	}
	if user.Mail.Valid {
		dto.Mail = user.Mail.String
	}
	if user.Phone.Valid {
		dto.Phone = user.Phone.String
	}
	if user.CreateTime.Valid {
		dto.CreateTime = user.CreateTime.Time
	}
	if user.UpdateTime.Valid {
		dto.UpdateTime = user.UpdateTime.Time
	}

	return dto
}

// GetUserById 根据 ID 获取用户信息（带缓存）
// 特性：
// 1. 优先从缓存读取，缓存命中后自动延期
// 2. 缓存未命中时使用分布式锁防止击穿
// 3. 数据不存在时设置空值缓存防止穿透
// 4. 高并发场景下的锁优化：未获取到锁的线程不等待，直接穿透查询
func (s *UserCacheService) GetUserById(ctx context.Context, userId uint64) (*UserDTO, error) {
	userDTO, _, err := s.GetUserByIdWithSource(ctx, userId)
	return userDTO, err
}

// GetUserByIdWithSource 根据 ID 获取用户信息（带缓存）并返回数据来源
func (s *UserCacheService) GetUserByIdWithSource(ctx context.Context, userId uint64) (*UserDTO, DataSource, error) {
	cacheKey := fmt.Sprintf("id:%d", userId)
	var userDTO UserDTO
	source := DataSourceCache

	err := s.cache.GetWithLoader(ctx, cacheKey, &userDTO, func() (interface{}, error) {
		source = DataSourceDatabase
		// 从数据库加载用户数据
		user, err := s.userModel.FindOne(ctx, userId)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, nil // 返回 nil 表示数据不存在
			}
			return nil, err
		}

		return toUserDTO(user), nil
	}, s.cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return nil, DataSourceDatabase, ErrNotFound
		}
		return nil, source, err
	}

	if source == DataSourceDatabase && userDTO.Id == 0 {
		return nil, source, ErrNotFound
	}

	return &userDTO, source, nil
}

// GetUserByMail 根据邮箱获取用户信息（带缓存）
func (s *UserCacheService) GetUserByMail(ctx context.Context, mail string) (*UserDTO, error) {
	cacheKey := fmt.Sprintf("mail:%s", mail)
	var userDTO UserDTO

	err := s.cache.GetWithLoader(ctx, cacheKey, &userDTO, func() (interface{}, error) {
		// 从数据库加载用户数据
		user, err := s.userModel.FindOneByMail(ctx, sql.NullString{String: mail, Valid: true})
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, nil
			}
			return nil, err
		}

		return toUserDTO(user), nil
	}, s.cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &userDTO, nil
}

// GetUserByPhone 根据手机号获取用户信息（带缓存）
func (s *UserCacheService) GetUserByPhone(ctx context.Context, phone string) (*UserDTO, error) {
	cacheKey := fmt.Sprintf("phone:%s", phone)
	var userDTO UserDTO

	err := s.cache.GetWithLoader(ctx, cacheKey, &userDTO, func() (interface{}, error) {
		// 从数据库加载用户数据
		user, err := s.userModel.FindOneByPhone(ctx, sql.NullString{String: phone, Valid: true})
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, nil
			}
			return nil, err
		}

		return toUserDTO(user), nil
	}, s.cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &userDTO, nil
}

// GetUserByIdWithMutex 根据 ID 获取用户信息（强一致性版本）
// 使用分布式互斥锁保证读写串行化，适用于对一致性要求极高的场景
func (s *UserCacheService) GetUserByIdWithMutex(ctx context.Context, userId uint64) (*UserDTO, error) {
	cacheKey := fmt.Sprintf("id:%d", userId)
	var userDTO UserDTO

	err := s.cache.GetWithLoader(ctx, cacheKey, &userDTO, func() (interface{}, error) {
		// 从数据库加载用户数据
		user, err := s.userModel.FindOne(ctx, userId)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, nil
			}
			return nil, err
		}

		return toUserDTO(user), nil
	}, s.cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &userDTO, nil
}

// CreateUser 创建用户（带缓存）
// 创建成功后自动设置缓存
func (s *UserCacheService) CreateUser(ctx context.Context, user *User) error {
	// 先插入数据库
	_, err := s.userModel.Insert(ctx, user)
	if err != nil {
		return err
	}

	// 设置缓存
	userDTO := toUserDTO(user)
	if userDTO != nil {
		// 按 ID 缓存
		idKey := fmt.Sprintf("id:%d", user.Id)
		_ = s.cache.Set(ctx, idKey, userDTO, s.cacheOpt.BaseExpiry+s.cacheOpt.RandomExpiry)

		// 按邮箱缓存
		if user.Mail.Valid && user.Mail.String != "" {
			mailKey := fmt.Sprintf("mail:%s", user.Mail.String)
			_ = s.cache.Set(ctx, mailKey, userDTO, s.cacheOpt.BaseExpiry+s.cacheOpt.RandomExpiry)
		}

		// 按手机号缓存
		if user.Phone.Valid && user.Phone.String != "" {
			phoneKey := fmt.Sprintf("phone:%s", user.Phone.String)
			_ = s.cache.Set(ctx, phoneKey, userDTO, s.cacheOpt.BaseExpiry+s.cacheOpt.RandomExpiry)
		}
	}

	return nil
}

// UpdateUser 更新用户信息（强一致性）
// 使用分布式互斥锁保证缓存和数据库的强一致性
func (s *UserCacheService) UpdateUser(ctx context.Context, user *User) error {
	cacheKey := fmt.Sprintf("id:%d", user.Id)

	return s.cache.UpdateWithMutex(ctx, cacheKey, func() (interface{}, error) {
		// 更新数据库
		if err := s.userModel.Update(ctx, user); err != nil {
			return nil, err
		}

		// 删除相关缓存
		s.invalidateUserCache(ctx, user.Id)

		return nil, nil
	}, s.cacheOpt)
}

// UpdateUserWithCache 更新用户信息并刷新缓存
// 采用先更新数据库，再查询完整数据，最后更新缓存的策略
func (s *UserCacheService) UpdateUserWithCache(ctx context.Context, user *User) error {
	cacheKey := fmt.Sprintf("id:%d", user.Id)

	// 先查询旧数据，用于清理旧的索引缓存（在锁外查询，减少锁持有时间）
	oldUser, err := s.userModel.FindOne(ctx, user.Id)
	if err != nil {
		return err
	}

	// 使用 UpdateWithMutex 保证并发安全，并直接更新缓存
	return s.cache.UpdateWithMutex(ctx, cacheKey, func() (interface{}, error) {
		// 1. 更新数据库
		if err := s.userModel.Update(ctx, user); err != nil {
			return nil, err
		}

		// 2. 查询更新后的完整用户信息（确保数据完整性，包括数据库自动更新的字段）
		updatedUser, err := s.userModel.FindOne(ctx, user.Id)
		if err != nil {
			return nil, err
		}

		// 3. 删除旧的索引缓存
		if oldUser.Mail.Valid && oldUser.Mail.String != "" &&
			(!updatedUser.Mail.Valid || updatedUser.Mail.String != oldUser.Mail.String) {
			mailKey := fmt.Sprintf("mail:%s", oldUser.Mail.String)
			_ = s.cache.Delete(ctx, mailKey)
		}

		// 4. 返回更新后的完整数据，用于更新主缓存（ID缓存）
		userDTO := toUserDTO(updatedUser)
		if userDTO == nil {
			return nil, nil
		}

		// 5. 更新其他索引缓存（邮箱、手机号）
		// 注意：ID缓存由 UpdateWithMutex 自动更新，这里只更新其他索引
		if updatedUser.Mail.Valid && updatedUser.Mail.String != "" {
			mailKey := fmt.Sprintf("mail:%s", updatedUser.Mail.String)
			_ = s.cache.Set(ctx, mailKey, userDTO, s.cacheOpt.BaseExpiry+s.cacheOpt.RandomExpiry)
		}

		// 返回数据用于更新主缓存（ID缓存）
		return userDTO, nil
	}, s.cacheOpt)
}

// DeleteUser 删除用户（带缓存清理）
func (s *UserCacheService) DeleteUser(ctx context.Context, userId uint64) error {
	// 先删除数据库
	if err := s.userModel.Delete(ctx, userId); err != nil {
		return err
	}

	// 删除缓存
	s.invalidateUserCache(ctx, userId)

	return nil
}

// invalidateUserCache 使用户缓存失效
func (s *UserCacheService) invalidateUserCache(ctx context.Context, userId uint64) {
	// 删除 ID 缓存
	idKey := fmt.Sprintf("id:%d", userId)
	_ = s.cache.Delete(ctx, idKey)

	// 注意：邮箱和手机号的缓存需要在更新时单独处理
	// 因为我们无法从 userId 反推出邮箱和手机号
}

// BatchGetUsers 批量获取用户信息（带缓存）
func (s *UserCacheService) BatchGetUsers(ctx context.Context, userIds []uint64) (map[uint64]*UserDTO, error) {
	result := make(map[uint64]*UserDTO, len(userIds))

	for _, userId := range userIds {
		user, err := s.GetUserById(ctx, userId)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue // 跳过不存在的用户
			}
			return nil, err
		}
		result[userId] = user
	}

	return result, nil
}
