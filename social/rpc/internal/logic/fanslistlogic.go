package logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/social/rpc/internal/svc"
	"github.com/hd2yao/ecshop/social/rpc/types/social"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type FansListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FansListLogic {
	return &FansListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FansList 粉丝列表（user_id 为 0 表示当前用户）
func (l *FansListLogic) FansList(in *social.FansListReq) (*social.ListResp, error) {
	// 1. 参数验证和默认值
	operatorId := in.OperatorId
	if operatorId <= 0 {
		return &social.ListResp{
			Code:    int32(errcode.UserTokenInvalid.Code()),
			Message: "用户未登录",
		}, nil
	}

	userId := in.UserId
	if userId == 0 {
		userId = operatorId
	}

	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.Size
	if size <= 0 {
		size = 20
	}

	l.Infof("查询粉丝列表，用户ID: %d, 页码: %d, 每页: %d", userId, page, size)

	// 2. 优先从 Redis list 读取
	cache := redisPool.NewRedisCache("social", "fans_list")
	fansListKey := fmt.Sprintf("fans:%d", userId)

	// 计算分页范围
	start := (page - 1) * size
	stop := start + size - 1

	// 从 Redis list 读取
	userIds, err := cache.LRange(l.ctx, fansListKey, start, stop)
	if err != nil {
		l.Errorf("从 Redis 读取粉丝列表失败: %v", err)
	}

	// 3. 如果 Redis 中没有数据，从数据库读取并写入 Redis
	if len(userIds) == 0 {
		offset := (page - 1) * size
		followers, total, err := l.svcCtx.UserFollowerModel.ListFollowers(l.ctx, userId, offset, size)
		if err != nil {
			l.Errorf("从数据库查询粉丝列表失败: %v", err)
			return &social.ListResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 写入 Redis list（从左侧插入，保持时间倒序）
		if len(followers) > 0 {
			// 先清空列表（如果存在），然后重新构建
			_ = cache.Delete(l.ctx, fansListKey)
			for i := len(followers) - 1; i >= 0; i-- {
				followerIdStr := fmt.Sprintf("%d", followers[i].FollowerId)
				_, _ = cache.LPush(l.ctx, fansListKey, followerIdStr)
			}
			// 重新读取当前页
			userIds, _ = cache.LRange(l.ctx, fansListKey, start, stop)
		} else {
			// 没有数据，返回空列表
			return &social.ListResp{
				Code:    int32(errcode.Success.Code()),
				Message: errcode.Success.Msg(),
				List:    []*social.UserBrief{},
				Total:   total,
				Page:    page,
				Size:    size,
			}, nil
		}
	}

	// 4. 获取总数（从 Redis list 长度或数据库）
	total, err := cache.LLen(l.ctx, fansListKey)
	if err != nil || total == 0 {
		// 如果 Redis 中没有总数，从数据库查询
		_, total, err = l.svcCtx.UserFollowerModel.ListFollowers(l.ctx, userId, 0, 1)
		if err != nil {
			total = int64(len(userIds))
		}
	}

	// 5. 批量获取用户信息并构建响应
	userBriefs := make([]*social.UserBrief, 0, len(userIds))
	for _, userIdStr := range userIds {
		followerId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			continue
		}

		// 调用 User RPC 获取用户信息
		userInfo, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoReq{
			UserId: followerId,
		})
		if err != nil || userInfo == nil || userInfo.UserInfo == nil {
			l.Errorf("获取用户信息失败，用户ID: %d, 错误: %v", followerId, err)
			continue
		}

		// 判断是否互相关注（需要查询当前用户是否关注了该粉丝）
		isFollow := false
		isMutual := false
		if operatorId > 0 {
			// 检查当前用户是否关注了该粉丝
			attention, err := l.svcCtx.UserAttentionModel.FindOneByUserAndAttention(l.ctx, operatorId, followerId)
			if err == nil && attention != nil && attention.IsDel == 0 {
				isFollow = true
			}
			// 检查是否互相关注（该粉丝是否也关注了当前用户）
			if isFollow {
				reverseAttention, err := l.svcCtx.UserAttentionModel.FindOneByUserAndAttention(l.ctx, followerId, operatorId)
				if err == nil && reverseAttention != nil && reverseAttention.IsDel == 0 {
					isMutual = true
				}
			}
		}

		userBriefs = append(userBriefs, &social.UserBrief{
			UserId:   followerId,
			Name:     userInfo.UserInfo.Name,
			Avatar:   userInfo.UserInfo.Avatar,
			IsFollow: isFollow,
			IsMutual: isMutual,
		})
	}

	// 6. 返回结果
	return &social.ListResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		List:    userBriefs,
		Total:   total,
		Page:    page,
		Size:    size,
	}, nil
}
