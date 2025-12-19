package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/rocketmq"
	"github.com/hd2yao/ecshop/social/model"
	"github.com/hd2yao/ecshop/social/rpc/internal/svc"
	"github.com/hd2yao/ecshop/social/rpc/types/social"
)

type FollowerUnfollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowerUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowerUnfollowLogic {
	return &FollowerUnfollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FollowerUnfollow 粉丝取关（取消对粉丝的关注）
func (l *FollowerUnfollowLogic) FollowerUnfollow(in *social.FollowReq) (*social.FollowResp, error) {
	// 1. 参数验证
	if in.OperatorId <= 0 {
		return &social.FollowResp{
			Code:    int32(errcode.UserTokenInvalid.Code()),
			Message: "用户未登录",
		}, nil
	}
	if in.UserId <= 0 {
		return &social.FollowResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "目标用户ID无效",
		}, nil
	}

	// 2. 获取操作者 ID
	operatorId := in.OperatorId

	// 3. 不能取关自己
	if operatorId == in.UserId {
		return &social.FollowResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "不能取关自己",
		}, nil
	}

	l.Infof("粉丝取关，粉丝：%d，博主：%d", operatorId, in.UserId)

	// 4. 检查是否已经关注过（从缓存中检查）
	exists, err := l.svcCtx.FollowCacheService.CheckFansListContains(l.ctx, in.UserId, operatorId)
	if err == nil && !exists {
		l.Infof("粉丝取关，粉丝 uid：%d 已经不存在列表中，不可以取关", operatorId)
		return &social.FollowResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "未关注该用户，无法取关",
		}, nil
	}

	// 5. 发布取关事件到 MQ
	event := model.FollowEvent{
		OperatorId: operatorId,
		TargetId:   in.UserId,
		Action:     model.FollowActionUnfollow,
		Role:       model.FollowRoleFollower,
	}

	rmq := rocketmq.GetRocketMQ()
	_, err = rmq.SendMessage(l.ctx, model.MQTopicFollow, model.FollowRoleFollower, event)
	if err != nil {
		l.Errorf("发布粉丝取关事件失败: %v", err)
		return &social.FollowResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	l.Infof("粉丝取关事件已发布，粉丝：%d，博主：%d", operatorId, in.UserId)

	return &social.FollowResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
	}, nil
}
