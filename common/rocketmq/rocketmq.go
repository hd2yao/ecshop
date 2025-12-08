package rocketmq

import (
	"context"
	"fmt"
	"sync"
)

// RocketMQ RocketMQ 客户端管理器
type RocketMQ struct {
	producer Producer
	consumer Consumer
	config   Config
	mu       sync.RWMutex
}

var (
	globalRocketMQ *RocketMQ
	rocketmqOnce   sync.Once
)

// InitRocketMQ 初始化 RocketMQ（单例模式）
func InitRocketMQ(config Config) error {
	var err error
	rocketmqOnce.Do(func() {
		// 验证配置
		if err = config.Validate(); err != nil {
			return
		}

		globalRocketMQ = &RocketMQ{
			config: config,
		}
	})

	return err
}

// GetRocketMQ 获取全局 RocketMQ 实例
func GetRocketMQ() *RocketMQ {
	if globalRocketMQ == nil {
		panic("RocketMQ 未初始化，请先调用 InitRocketMQ()")
	}
	return globalRocketMQ
}

// GetProducer 获取生产者（懒加载）
func (r *RocketMQ) GetProducer() (Producer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.producer == nil {
		producer, err := NewProducer(r.config)
		if err != nil {
			return nil, fmt.Errorf("创建生产者失败: %w", err)
		}

		if err := producer.Start(); err != nil {
			return nil, fmt.Errorf("启动生产者失败: %w", err)
		}

		r.producer = producer
	}

	return r.producer, nil
}

// GetConsumer 获取消费者（懒加载）
func (r *RocketMQ) GetConsumer() (Consumer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.consumer == nil {
		consumer, err := NewConsumer(r.config)
		if err != nil {
			return nil, fmt.Errorf("创建消费者失败: %w", err)
		}

		r.consumer = consumer
	}

	return r.consumer, nil
}

// SendMessage 发送消息（便捷方法）
func (r *RocketMQ) SendMessage(ctx context.Context, topic, tag string, body interface{}) (*SendResult, error) {
	producer, err := r.GetProducer()
	if err != nil {
		return nil, err
	}

	msg, err := NewMessage(topic, tag, body)
	if err != nil {
		return nil, err
	}

	return producer.Send(ctx, msg)
}

// SendEventMessage 发送事件消息（便捷方法）
func (r *RocketMQ) SendEventMessage(ctx context.Context, topic, tag string, event *EventMessage) (*SendResult, error) {
	producer, err := r.GetProducer()
	if err != nil {
		return nil, err
	}

	msg, err := event.ToRocketMQMessage(topic, tag)
	if err != nil {
		return nil, err
	}

	return producer.Send(ctx, msg)
}

// SendMessageAsync 异步发送消息（便捷方法）
func (r *RocketMQ) SendMessageAsync(ctx context.Context, topic, tag string, body interface{}, callback func(*SendResult, error)) error {
	producer, err := r.GetProducer()
	if err != nil {
		return err
	}

	msg, err := NewMessage(topic, tag, body)
	if err != nil {
		if callback != nil {
			callback(nil, err)
		}
		return err
	}

	return producer.SendAsync(ctx, msg, callback)
}

// Subscribe 订阅主题（便捷方法）
func (r *RocketMQ) Subscribe(topic string, selector MessageSelector, listener MessageListener) error {
	consumer, err := r.GetConsumer()
	if err != nil {
		return err
	}

	return consumer.Subscribe(topic, selector, listener)
}

// StartConsumer 启动消费者（便捷方法）
func (r *RocketMQ) StartConsumer() error {
	consumer, err := r.GetConsumer()
	if err != nil {
		return err
	}

	return consumer.Start()
}

// Shutdown 关闭所有连接
func (r *RocketMQ) Shutdown() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error

	if r.producer != nil {
		if err := r.producer.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("关闭生产者失败: %w", err))
		}
		r.producer = nil
	}

	if r.consumer != nil {
		if err := r.consumer.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("关闭消费者失败: %w", err))
		}
		r.consumer = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭 RocketMQ 连接时发生错误: %v", errs)
	}

	return nil
}

