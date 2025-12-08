package rocketmq

import (
	"encoding/json"
	"fmt"
	"time"
)

// Message RocketMQ 消息结构
type Message struct {
	// Topic 主题
	Topic string `json:"topic"`

	// Tag 标签（用于消息过滤）
	Tag string `json:"tag"`

	// Key 消息Key（用于消息去重和查询）
	Key string `json:"key"`

	// Body 消息体（JSON 字符串或字节数组）
	Body []byte `json:"body"`

	// Properties 消息属性（扩展字段）
	Properties map[string]string `json:"properties"`

	// DelayTimeLevel 延迟消息级别（1-18，对应不同延迟时间）
	// 0 表示不延迟
	DelayTimeLevel int `json:"delay_time_level"`

	// 内部字段：消息ID（发送后由 RocketMQ 生成）
	MessageID string `json:"message_id,omitempty"`

	// 内部字段：发送时间
	SendTime time.Time `json:"send_time,omitempty"`
}

// NewMessage 创建新消息
func NewMessage(topic, tag string, body interface{}) (*Message, error) {
	var bodyBytes []byte
	var err error

	switch v := body.(type) {
	case []byte:
		bodyBytes = v
	case string:
		bodyBytes = []byte(v)
	default:
		// 尝试序列化为 JSON
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化消息体失败: %w", err)
		}
	}

	return &Message{
		Topic:      topic,
		Tag:        tag,
		Body:       bodyBytes,
		Properties: make(map[string]string),
	}, nil
}

// SetKey 设置消息Key
func (m *Message) SetKey(key string) *Message {
	m.Key = key
	return m
}

// SetProperty 设置消息属性
func (m *Message) SetProperty(key, value string) *Message {
	if m.Properties == nil {
		m.Properties = make(map[string]string)
	}
	m.Properties[key] = value
	return m
}

// SetDelayTimeLevel 设置延迟消息级别
// 级别对应关系：
// 1: 1s, 2: 5s, 3: 10s, 4: 30s, 5: 1m, 6: 2m, 7: 3m, 8: 4m, 9: 5m,
// 10: 6m, 11: 7m, 12: 8m, 13: 9m, 14: 10m, 15: 20m, 16: 30m, 17: 1h, 18: 2h
func (m *Message) SetDelayTimeLevel(level int) *Message {
	if level < 0 || level > 18 {
		level = 0
	}
	m.DelayTimeLevel = level
	return m
}

// GetBodyString 获取消息体字符串
func (m *Message) GetBodyString() string {
	return string(m.Body)
}

// UnmarshalBody 反序列化消息体到指定结构
func (m *Message) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(m.Body, v)
}

// EventMessage 事件消息（用于业务层）
type EventMessage struct {
	// EventType 事件类型
	EventType string `json:"event_type"`

	// UserID 用户ID
	UserID int64 `json:"user_id"`

	// FoodID 美食ID（可选）
	FoodID int64 `json:"food_id,omitempty"`

	// Timestamp 时间戳
	Timestamp int64 `json:"timestamp"`

	// Data 扩展数据
	Data map[string]interface{} `json:"data,omitempty"`
}

// NewEventMessage 创建事件消息
func NewEventMessage(eventType string, userID int64, foodID ...int64) *EventMessage {
	msg := &EventMessage{
		EventType: eventType,
		UserID:    userID,
		Timestamp: time.Now().Unix(),
		Data:      make(map[string]interface{}),
	}
	if len(foodID) > 0 {
		msg.FoodID = foodID[0]
	}
	return msg
}

// SetData 设置扩展数据
func (e *EventMessage) SetData(key string, value interface{}) *EventMessage {
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
	e.Data[key] = value
	return e
}

// ToRocketMQMessage 转换为 RocketMQ 消息
func (e *EventMessage) ToRocketMQMessage(topic, tag string) (*Message, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("序列化事件消息失败: %w", err)
	}

	msg, err := NewMessage(topic, tag, body)
	if err != nil {
		return nil, err
	}

	// 设置消息Key（用于去重和查询）
	msg.SetKey(fmt.Sprintf("%s_%d_%d", e.EventType, e.UserID, e.FoodID))

	// 设置消息属性
	msg.SetProperty("event_type", e.EventType)
	msg.SetProperty("user_id", fmt.Sprintf("%d", e.UserID))
	if e.FoodID > 0 {
		msg.SetProperty("food_id", fmt.Sprintf("%d", e.FoodID))
	}

	return msg, nil
}

