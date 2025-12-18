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

type FollowListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowListLogic {
	return &FollowListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FollowList 关注列表
func (l *FollowListLogic) FollowList(in *social.FollowListReq) (*social.ListResp, error) {
	// 1. 参数验证：必须有操作者ID
	userId := in.OperatorId
	if userId <= 0 {
		return &social.ListResp{
			Code:    int32(errcode.UserTokenInvalid.Code()),
			Message: "用户未登录",
		}, nil
	}

	// 2. 参数验证和默认值
	page := in.Page
	if page <= 0 {
		page = 1
	}
	size := in.Size
	if size <= 0 {
		size = 20
	}

	l.Infof("查询关注列表，用户ID: %d, 页码: %d, 每页: %d", userId, page, size)

	// 3. 优先从 Redis list 读取
	cache := redisPool.NewRedisCache("social", "follow_list")
	followListKey := fmt.Sprintf("follow:%d", userId)

	// 计算分页范围
	start := (page - 1) * size
	stop := start + size - 1

	// 从 Redis list 读取
	userIds, err := cache.LRange(l.ctx, followListKey, start, stop)
	if err != nil {
		l.Errorf("从 Redis 读取关注列表失败: %v", err)
	}

	// 4. 如果 Redis 中没有数据，从数据库读取并写入 Redis
	if len(userIds) == 0 {
		offset := (page - 1) * size
		attentions, total, err := l.svcCtx.UserAttentionModel.ListAttentions(l.ctx, userId, offset, size)
		if err != nil {
			l.Errorf("从数据库查询关注列表失败: %v", err)
			return &social.ListResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 写入 Redis list（从左侧插入，保持时间倒序）
		if len(attentions) > 0 {
			// 先清空列表（如果存在），然后重新构建
			_ = cache.Delete(l.ctx, followListKey)
			for i := len(attentions) - 1; i >= 0; i-- {
				attentionIdStr := fmt.Sprintf("%d", attentions[i].AttentionId)
				_, _ = cache.LPush(l.ctx, followListKey, attentionIdStr)
			}
			// 重新读取当前页
			userIds, _ = cache.LRange(l.ctx, followListKey, start, stop)
		} else {
			// 没有数据，返回空列表
			return &social.ListResp{
				Code:    int32(errcode.Success.Code()),
				Message: errcode.Success.Msg(),
				List:    []*social.UserBrief{},
				Total:   0,
				Page:    page,
				Size:    size,
			}, nil
		}
	}

	// 5. 获取总数（从 Redis list 长度或数据库）
	total, err := cache.LLen(l.ctx, followListKey)
	if err != nil || total == 0 {
		// 如果 Redis 中没有总数，从数据库查询
		_, total, err = l.svcCtx.UserAttentionModel.ListAttentions(l.ctx, userId, 0, 1)
		if err != nil {
			total = int64(len(userIds))
		}
	}

	// 6. 批量获取用户信息并构建响应
	userBriefs := make([]*social.UserBrief, 0, len(userIds))
	for _, userIdStr := range userIds {
		attentionId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			continue
		}

		// 调用 User RPC 获取用户信息
		userInfo, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.GetUserInfoReq{
			UserId: attentionId,
		})
		if err != nil || userInfo == nil || userInfo.UserInfo == nil {
			l.Errorf("获取用户信息失败，用户ID: %d, 错误: %v", attentionId, err)
			continue
		}

		// 判断是否互相关注（当前用户关注了该用户，且该用户也关注了当前用户）
		isMutual := false
		reverseAttention, err := l.svcCtx.UserAttentionModel.FindOneByUserAndAttention(l.ctx, attentionId, userId)
		if err == nil && reverseAttention != nil && reverseAttention.IsDel == 0 {
			isMutual = true
		}

		userBriefs = append(userBriefs, &social.UserBrief{
			UserId:   attentionId,
			Name:     userInfo.UserInfo.Name,
			Avatar:   userInfo.UserInfo.Avatar,
			IsFollow: true, // 关注列表中都是已关注的
			IsMutual: isMutual,
		})
	}

	// 7. 返回结果
	return &social.ListResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		List:    userBriefs,
		Total:   total,
		Page:    page,
		Size:    size,
	}, nil
}
