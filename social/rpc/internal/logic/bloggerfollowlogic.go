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

type BloggerFollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBloggerFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BloggerFollowLogic {
	return &BloggerFollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BloggerFollow 博主关注
func (l *BloggerFollowLogic) BloggerFollow(in *social.FollowReq) (*social.FollowResp, error) {
	// 1. 参数验证
	if in.OperatorId <= 0 {
		return &social.FollowResp{
			Code:    int32(errcode.UserTokenInvalid.Code()),
			Message: "用户未登录",
		}, nil
	}
	if in.UserId <= 0 {
		return &social.FollowResp{
			Code:    int32(errcode.SocialTargetUserIdInvalid.Code()),
			Message: errcode.SocialTargetUserIdInvalid.Msg(),
		}, nil
	}

	// 2. 获取操作者 ID
	operatorId := in.OperatorId

	// 3. 不能关注自己
	if in.OperatorId == in.UserId {
		return &social.FollowResp{
			Code:    int32(errcode.SocialFollowSelfError.Code()),
			Message: errcode.SocialFollowSelfError.Msg(),
		}, nil
	}

	l.Infof("博主关注，当前博主：%d，其他被关注博主：%d", operatorId, in.UserId)

	// 4. 检查是否已经关注过（从缓存中检查）
	exists, err := l.svcCtx.FollowModel.CacheService().CheckFollowListContains(l.ctx, operatorId, in.UserId)
	if err == nil && exists {
		l.Infof("博主关注，其他被关注博主 uid：%d 已经存在列表中", in.UserId)
		return &social.FollowResp{
			Code:    int32(errcode.SocialFollowAlreadyExists.Code()),
			Message: errcode.SocialFollowAlreadyExists.Msg(),
		}, nil
	}

	// 5. 发布关注事件到 MQ
	event := model.FollowEvent{
		OperatorId: operatorId,
		TargetId:   in.UserId,
		Action:     model.FollowActionFollow,
		Role:       model.FollowRoleBlogger,
	}

	rmq := rocketmq.GetRocketMQ()
	_, err = rmq.SendMessage(l.ctx, model.MQTopicFollow, model.FollowRoleBlogger, event)
	if err != nil {
		l.Errorf("发布博主关注事件失败: %v", err)
		return &social.FollowResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	l.Infof("博主关注事件已发布，当前博主：%d，其他被关注博主：%d", operatorId, in.UserId)

	return &social.FollowResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
	}, nil
}
