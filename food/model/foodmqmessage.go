package model

// FoodUpdateMQMessage 美食更新消息（用于 RocketMQ）
type FoodUpdateMQMessage struct {
	// FoodID 美食ID
	FoodID int64 `json:"food_id"`

	// UserID 用户ID
	UserID int64 `json:"user_id"`
}

