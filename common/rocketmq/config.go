package rocketmq

import "fmt"

// Config RocketMQ 配置
type Config struct {
	// NameServer 地址（多个用分号分隔）
	// 示例: "127.0.0.1:9876" 或 "127.0.0.1:9876;127.0.0.2:9876"
	NameServer string `json:"name_server"`

	// 生产者配置
	Producer ProducerConfig `json:"producer"`

	// 消费者配置
	Consumer ConsumerConfig `json:"consumer"`
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	// 生产者组名
	GroupName string `json:"group_name"`

	// 发送消息超时时间（毫秒），默认 3000
	SendMsgTimeout int `json:"send_msg_timeout"`

	// 消息最大大小（字节），默认 4MB
	MaxMessageSize int `json:"max_message_size"`

	// 压缩消息体阈值（字节），超过此值将压缩，默认 4096
	CompressMsgBodyOverHowmuch int `json:"compress_msg_body_over_howmuch"`

	// 压缩级别（1-9），默认 5
	CompressLevel int `json:"compress_level"`

	// 重试次数，默认 2
	RetryTimes int `json:"retry_times"`

	// 是否启用消息轨迹，默认 false
	EnableMsgTrace bool `json:"enable_msg_trace"`

	// 是否启用 VIP 通道，默认 false
	VipChannelEnabled bool `json:"vip_channel_enabled"`
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	// 消费者组名
	GroupName string `json:"group_name"`

	// 消费模式：BROADCASTING（广播）或 CLUSTERING（集群），默认 CLUSTERING
	MessageModel string `json:"message_model"`

	// 消费类型：CONSUME_ACTIVELY（主动拉取）或 CONSUME_PASSIVELY（被动推送），默认 CONSUME_PASSIVELY
	ConsumeFromWhere string `json:"consume_from_where"`

	// 消费线程数，默认 20
	ConsumeThreadMin int `json:"consume_thread_min"`

	// 消费线程最大数，默认 20
	ConsumeThreadMax int `json:"consume_thread_max"`

	// 批量消费消息数量，默认 1
	ConsumeMessageBatchMaxSize int `json:"consume_message_batch_max_size"`

	// 最大重试次数，默认 16
	MaxReconsumeTimes int `json:"max_reconsume_times"`

	// 是否启用消息轨迹，默认 false
	EnableMsgTrace bool `json:"enable_msg_trace"`

	// 是否启用 VIP 通道，默认 false
	VipChannelEnabled bool `json:"vip_channel_enabled"`
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		NameServer: "127.0.0.1:9876",
		Producer: ProducerConfig{
			GroupName:                  "default_producer_group",
			SendMsgTimeout:             3000,
			MaxMessageSize:             4 * 1024 * 1024, // 4MB
			CompressMsgBodyOverHowmuch: 4096,
			CompressLevel:              5,
			RetryTimes:                 2,
			EnableMsgTrace:             false,
			VipChannelEnabled:          false,
		},
		Consumer: ConsumerConfig{
			GroupName:                  "default_consumer_group",
			MessageModel:               "CLUSTERING",
			ConsumeFromWhere:           "CONSUME_FROM_LAST_OFFSET",
			ConsumeThreadMin:           20,
			ConsumeThreadMax:           20,
			ConsumeMessageBatchMaxSize: 1,
			MaxReconsumeTimes:          16,
			EnableMsgTrace:             false,
			VipChannelEnabled:          false,
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.NameServer == "" {
		return fmt.Errorf("name_server is required")
	}
	if c.Producer.GroupName == "" {
		return fmt.Errorf("producer.group_name is required")
	}
	if c.Consumer.GroupName == "" {
		return fmt.Errorf("consumer.group_name is required")
	}
	return nil
}

// Validate 验证生产者配置
func (p *ProducerConfig) Validate() error {
	if p.GroupName == "" {
		return fmt.Errorf("producer group_name is required")
	}
	if p.SendMsgTimeout <= 0 {
		p.SendMsgTimeout = 3000
	}
	if p.MaxMessageSize <= 0 {
		p.MaxMessageSize = 4 * 1024 * 1024 // 4MB
	}
	if p.RetryTimes < 0 {
		p.RetryTimes = 2
	}
	return nil
}

// Validate 验证消费者配置
func (c *ConsumerConfig) Validate() error {
	if c.GroupName == "" {
		return fmt.Errorf("consumer group_name is required")
	}
	if c.ConsumeThreadMin <= 0 {
		c.ConsumeThreadMin = 20
	}
	if c.ConsumeThreadMax <= 0 {
		c.ConsumeThreadMax = 20
	}
	if c.ConsumeMessageBatchMaxSize <= 0 {
		c.ConsumeMessageBatchMaxSize = 1
	}
	if c.MaxReconsumeTimes < 0 {
		c.MaxReconsumeTimes = 16
	}
	return nil
}
