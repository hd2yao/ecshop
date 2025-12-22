package model

// RocketMQ 主题与标签定义（美食缓存更新）
const (
	// MQTopicFoodUpdate 美食缓存更新主题
	MQTopicFoodUpdate = "food_cache_update"
	
	FoodTagCreate = "create"
	FoodTagUpdate = "update"
	FoodTagDelete = "delete"
	FoodTagInit   = "init"
)

// FoodUpdateMQMessage 美食更新消息（用于 RocketMQ）
type FoodUpdateMQMessage struct {
	// FoodID 美食ID
	FoodID int64 `json:"food_id"`

	// UserID 用户ID
	UserID int64 `json:"user_id"`
}
