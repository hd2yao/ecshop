package model

// 关注/取关事件定义
const (
	MQTopicFollow = "social_follow"

	FollowActionFollow   = "follow"
	FollowActionUnfollow = "unfollow"

	FollowRoleBlogger  = "blogger"
	FollowRoleFollower = "follower"
)

// FollowEvent MQ 消息体
type FollowEvent struct {
	OperatorId int64  `json:"operator_id"` // 发起操作的用户
	TargetId   int64  `json:"target_id"`   // 目标用户
	Action     string `json:"action"`      // follow / unfollow
	Role       string `json:"role"`        // blogger / follower
}

