package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
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

	// 2. 从缓存服务获取粉丝列表
	followerIds, total, err := l.svcCtx.FollowCacheService.GetFansList(l.ctx, userId, int32(page), int32(size))
	if err != nil {
		l.Errorf("获取粉丝列表失败: %v", err)
		return &social.ListResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 如果没有数据，返回空列表
	if len(followerIds) == 0 {
		return &social.ListResp{
			Code:    int32(errcode.Success.Code()),
			Message: errcode.Success.Msg(),
			List:    []*social.UserBrief{},
			Total:   total,
			Page:    page,
			Size:    size,
		}, nil
	}

	// 3. 批量获取用户信息并构建响应
	userBriefs := make([]*social.UserBrief, 0, len(followerIds))
	for _, followerId := range followerIds {

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

	// 4. 返回结果
	return &social.ListResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		List:    userBriefs,
		Total:   total,
		Page:    page,
		Size:    size,
	}, nil
}
