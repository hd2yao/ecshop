package svc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/mail"
	"github.com/hd2yao/ecshop/common/oss"
	redisPool "github.com/hd2yao/ecshop/common/redis"
	socialModel "github.com/hd2yao/ecshop/social/model"
	"github.com/hd2yao/ecshop/user/model"
	"github.com/hd2yao/ecshop/user/rpc/internal/config"
)

type ServiceContext struct {
	Config             config.Config
	MailService        *mail.MailService
	OssClient          *oss.Client
	UserModel          model.UserModel
	UserAddressModel   model.UserAddressModel
	UserRelationModel  socialModel.UserRelationModel
	UserCache          *redisPool.RedisCache
	RefreshTokenCache  *redisPool.RedisCache
	PreviewAvatarCache *redisPool.RedisCache
	FollowStatCache    *redisPool.RedisCache
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Redis 连接池（从 CacheRedis 配置中提取第一个 Redis 配置）
	if len(c.CacheRedis) > 0 {
		redisConf := c.CacheRedis[0]
		if err := redisPool.InitRedisPoolFromGoZero(redisConf.RedisConf); err != nil {
			log.Fatalf("初始化Redis连接池失败: %v", err)
		}
	}

	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DataSource)

	// 初始化 Redis 缓存（需要在创建 UserModel 之前初始化）
	userCache := redisPool.NewRedisCache("user", "info")

	// 初始化用户模型（同时用于数据缓存和业务数据存储）
	userModel := model.NewUserModel(conn, c.CacheRedis, userCache)

	// 初始化用户地址模型
	userAddressModel := model.NewUserAddressModel(conn, c.CacheRedis)

	// 初始化用户关系模型（用于查询关注数、粉丝数）
	userRelationModel := socialModel.NewUserRelationModel(conn, c.CacheRedis)

	// 初始化邮件服务 - 从环境变量读取配置
	mailConfig := c.Mail
	if username := os.Getenv("MAIL_USERNAME"); username != "" {
		mailConfig.SMTP.Username = username
	}
	if password := os.Getenv("MAIL_PASSWORD"); password != "" {
		mailConfig.SMTP.Password = password
	}
	if from := os.Getenv("MAIL_FROM"); from != "" {
		mailConfig.SMTP.From = from
	}

	mailService, err := mail.NewMailService(mailConfig, "user", 10*time.Minute)
	if err != nil {
		log.Fatalf("初始化邮件服务失败: %v", err)
	}

	// 初始化 OSS 客户端
	ossConfig := c.Oss
	if accessKey := os.Getenv("OSS_ACCESS_KEY_ID"); accessKey != "" {
		ossConfig.AccessKeyId = accessKey
	}
	if accessSecret := os.Getenv("OSS_ACCESS_KEY_SECRET"); accessSecret != "" {
		ossConfig.AccessKeySecret = accessSecret
	}

	ossClient, err := oss.NewClient(&ossConfig)
	if err != nil {
		log.Fatalf("初始化 OSS 客户端失败: %v", err)
	}

	// 初始化其他 Redis 缓存
	refreshTokenCache := redisPool.NewRedisCache("user", "refresh_token")
	previewAvatarCache := redisPool.NewRedisCache("user", "preview_avatar")
	followStatCache := redisPool.NewRedisCache("user", "follow_stat")

	return &ServiceContext{
		Config:             c,
		MailService:        mailService,
		OssClient:          ossClient,
		UserModel:          userModel,
		UserAddressModel:   userAddressModel,
		UserRelationModel:  userRelationModel,
		UserCache:          userCache,
		RefreshTokenCache:  refreshTokenCache,
		PreviewAvatarCache: previewAvatarCache,
		FollowStatCache:    followStatCache,
	}
}

// GetFollowStat 获取用户的关注数和粉丝数（带缓存）
// 第一次从数据库读取，下次直接读缓存了，如果是热点用户会自动续期的
func (s *ServiceContext) GetFollowStat(ctx context.Context, userId int64) (followCount, followerCount int64, err error) {
	cacheOpt := redisPool.DefaultCacheOption()
	cacheKey := fmt.Sprintf("user_id:%d", userId)

	type FollowStat struct {
		FollowCount   int64 `json:"follow_count"`
		FollowerCount int64 `json:"follower_count"`
	}

	var stat FollowStat
	err = s.FollowStatCache.GetWithLoader(ctx, cacheKey, &stat, func() (interface{}, error) {
		// 从数据库加载关注统计数据
		relation, err := s.UserRelationModel.FindOneByUserId(ctx, userId)
		if err != nil {
			if errors.Is(err, socialModel.ErrNotFound) {
				// 用户没有关注关系记录，返回默认值
				return &FollowStat{
					FollowCount:   0,
					FollowerCount: 0,
				}, nil
			}
			return nil, err
		}

		return &FollowStat{
			FollowCount:   relation.AttentionCount,
			FollowerCount: relation.FollowerCount,
		}, nil
	}, cacheOpt)

	if err != nil {
		if errors.Is(err, errcode.CacheNotFound) {
			// 缓存未找到，返回默认值
			return 0, 0, nil
		}
		return 0, 0, err
	}

	return stat.FollowCount, stat.FollowerCount, nil
}
