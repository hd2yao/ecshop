package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
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

	// 3. 从缓存服务获取关注列表
	attentionIds, total, err := l.svcCtx.FollowCacheService.GetFollowList(l.ctx, userId, int32(page), int32(size))
	if err != nil {
		l.Errorf("获取关注列表失败: %v", err)
		return &social.ListResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 如果没有数据，返回空列表
	if len(attentionIds) == 0 {
		return &social.ListResp{
			Code:    int32(errcode.Success.Code()),
			Message: errcode.Success.Msg(),
			List:    []*social.UserBrief{},
			Total:   total,
			Page:    page,
			Size:    size,
		}, nil
	}

	// 4. 批量获取用户信息并构建响应
	userBriefs := make([]*social.UserBrief, 0, len(attentionIds))
	for _, attentionId := range attentionIds {

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

	// 5. 返回结果
	return &social.ListResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		List:    userBriefs,
		Total:   total,
		Page:    page,
		Size:    size,
	}, nil
}
