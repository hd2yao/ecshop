package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/rocketmq"
	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/social/rpc/internal/svc"
	"github.com/hd2yao/ecshop/social/rpc/types/social"
)

type FollowerFollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowerFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowerFollowLogic {
	return &FollowerFollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FollowerFollow 粉丝关注（回关粉丝）
func (l *FollowerFollowLogic) FollowerFollow(in *social.FollowReq) (*social.FollowResp, error) {
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

	// 3. 不能关注自己
	if operatorId == in.UserId {
		return &social.FollowResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "不能关注自己",
		}, nil
	}

	l.Infof("粉丝关注，粉丝：%d，博主：%d", operatorId, in.UserId)

	// 4. 检查是否已经关注过（从 Redis list 中检查）
	cache := redisPool.NewRedisCache("social", "fans_list")
	fansListKey := fmt.Sprintf("fans:%d", in.UserId)
	values, err := cache.LRange(l.ctx, fansListKey, 0, -1)
	if err == nil {
		operatorIdStr := fmt.Sprintf("%d", operatorId)
		for _, v := range values {
			if v == operatorIdStr {
				l.Infof("粉丝关注，粉丝 uid：%d 已经存在列表中", operatorId)
				return &social.FollowResp{
					Code:    int32(errcode.CommonParamError.Code()),
					Message: "已经关注过该用户",
				}, nil
			}
		}
	}

	// 5. 发布关注事件到 MQ
	event := FollowEvent{
		OperatorId: operatorId,
		TargetId:   in.UserId,
		Action:     FollowActionFollow,
		Role:       FollowRoleFollower,
	}

	rmq := rocketmq.GetRocketMQ()
	_, err = rmq.SendMessage(l.ctx, MQTopicFollow, FollowRoleFollower, event)
	if err != nil {
		l.Errorf("发布粉丝关注事件失败: %v", err)
		return &social.FollowResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	l.Infof("粉丝关注事件已发布，粉丝：%d，博主：%d", operatorId, in.UserId)

	return &social.FollowResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
	}, nil
}
