package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	redisPool "github.com/hd2yao/ecshop/common/redis"
	"github.com/hd2yao/ecshop/common/rocketmq"
	"github.com/hd2yao/ecshop/social/rpc/internal/svc"
)

// FollowEventListener 关注/取关事件监听器
type FollowEventListener struct {
	svcCtx *svc.ServiceContext
	logger logx.Logger
}

// NewFollowEventListener 创建关注/取关事件监听器
func NewFollowEventListener(svcCtx *svc.ServiceContext) *FollowEventListener {
	return &FollowEventListener{
		svcCtx: svcCtx,
		logger: logx.WithContext(context.Background()),
	}
}

// ConsumeMessage 消费消息
func (l *FollowEventListener) ConsumeMessage(ctx context.Context, msgs []*rocketmq.Message) (rocketmq.ConsumeStatus, error) {
	l.logger.Infof("关注/取关 MQ 消费消息-收到 %d 条消息", len(msgs))

	for i, msg := range msgs {
		l.logger.Infof("关注/取关 MQ 消费消息-第 %d/%d 条: MessageID=%s, Topic=%s, Tag=%s, Key=%s",
			i+1, len(msgs), msg.MessageID, msg.Topic, msg.Tag, msg.Key)

		// 解析消息
		var event FollowEvent
		if err := msg.UnmarshalBody(&event); err != nil {
			l.logger.Errorf("关注/取关 MQ 消费消息-解析消息体失败: MessageID=%s, error=%v, body=%s",
				msg.MessageID, err, msg.GetBodyString())
			continue
		}

		l.logger.Infof("关注/取关 MQ 消费消息-开始处理: operatorId=%d, targetId=%d, role=%s, action=%s",
			event.OperatorId, event.TargetId, event.Role, event.Action)

		// 处理事件
		if err := l.handleEvent(ctx, &event); err != nil {
			l.logger.Errorf("关注/取关 MQ 消费消息-处理失败: operatorId=%d, targetId=%d, error=%v",
				event.OperatorId, event.TargetId, err)
			return rocketmq.ConsumeLater, err
		}

		l.logger.Infof("关注/取关 MQ 消费消息-处理成功: operatorId=%d, targetId=%d", event.OperatorId, event.TargetId)
	}

	l.logger.Infof("关注/取关 MQ 消费消息-批量处理完成: 共 %d 条消息", len(msgs))
	return rocketmq.ConsumeSuccess, nil
}

// handleEvent 落库并维护缓存
func (l *FollowEventListener) handleEvent(ctx context.Context, e *FollowEvent) error {
	switch e.Role {
	case FollowRoleBlogger:
		if e.Action == FollowActionFollow {
			if err := l.svcCtx.UserAttentionModel.UpsertAttention(ctx, e.OperatorId, e.TargetId); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowCount(ctx, e.OperatorId, 1); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowerCount(ctx, e.TargetId, 1); err != nil {
				return err
			}
			// 缓存：当前用户关注列表头插
			cache := redisPool.NewRedisCache("social", "follow_list")
			_, _ = cache.LPush(ctx, followListKey(e.OperatorId), idToStr(e.TargetId))
			// 缓存：目标用户粉丝列表头插
			cacheFan := redisPool.NewRedisCache("social", "fans_list")
			_, _ = cacheFan.LPush(ctx, fansListKey(e.TargetId), idToStr(e.OperatorId))
		} else {
			if err := l.svcCtx.UserAttentionModel.SoftDeleteAttention(ctx, e.OperatorId, e.TargetId); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowCount(ctx, e.OperatorId, -1); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowerCount(ctx, e.TargetId, -1); err != nil {
				return err
			}
			// 取关时删除缓存，交给下次读重建
			cache := redisPool.NewRedisCache("social", "follow_list")
			_ = cache.Delete(ctx, followListKey(e.OperatorId))
			cacheFan := redisPool.NewRedisCache("social", "fans_list")
			_ = cacheFan.Delete(ctx, fansListKey(e.TargetId))
		}
	case FollowRoleFollower:
		if e.Action == FollowActionFollow {
			if err := l.svcCtx.UserFollowerModel.UpsertFollower(ctx, e.TargetId, e.OperatorId); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowerCount(ctx, e.TargetId, 1); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowCount(ctx, e.OperatorId, 1); err != nil {
				return err
			}
			cacheFan := redisPool.NewRedisCache("social", "fans_list")
			_, _ = cacheFan.LPush(ctx, fansListKey(e.TargetId), idToStr(e.OperatorId))
			cacheFollow := redisPool.NewRedisCache("social", "follow_list")
			_, _ = cacheFollow.LPush(ctx, followListKey(e.OperatorId), idToStr(e.TargetId))
		} else {
			if err := l.svcCtx.UserFollowerModel.SoftDeleteFollower(ctx, e.TargetId, e.OperatorId); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowerCount(ctx, e.TargetId, -1); err != nil {
				return err
			}
			if err := l.svcCtx.UserRelationModel.AddFollowCount(ctx, e.OperatorId, -1); err != nil {
				return err
			}
			cacheFan := redisPool.NewRedisCache("social", "fans_list")
			_ = cacheFan.Delete(ctx, fansListKey(e.TargetId))
			cacheFollow := redisPool.NewRedisCache("social", "follow_list")
			_ = cacheFollow.Delete(ctx, followListKey(e.OperatorId))
		}
	default:
		l.logger.Errorf("未知角色: %s", e.Role)
	}
	return nil
}

func followListKey(userId int64) string {
	return fmt.Sprintf("follow:%d", userId)
}

func fansListKey(userId int64) string {
	return fmt.Sprintf("fans:%d", userId)
}

func idToStr(id int64) string {
	return fmt.Sprintf("%d", id)
}
