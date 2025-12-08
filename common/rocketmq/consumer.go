package rocketmq

import (
	"context"
	"fmt"
	"strings"
	"sync"

	rmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"
)

// MessageListener 消息监听器接口
type MessageListener interface {
	// ConsumeMessage 消费消息
	// 返回值：
	//   - ConsumeStatus: 消费状态（SUCCESS 或 RECONSUME_LATER）
	//   - error: 错误信息
	ConsumeMessage(ctx context.Context, msgs []*Message) (ConsumeStatus, error)
}

// ConsumeStatus 消费状态
type ConsumeStatus int

const (
	// ConsumeSuccess 消费成功
	ConsumeSuccess ConsumeStatus = iota
	// ConsumeLater 稍后重新消费
	ConsumeLater
)

// String 返回状态字符串
func (s ConsumeStatus) String() string {
	switch s {
	case ConsumeSuccess:
		return "SUCCESS"
	case ConsumeLater:
		return "RECONSUME_LATER"
	default:
		return "UNKNOWN"
	}
}

// Consumer RocketMQ 消费者接口
type Consumer interface {
	// Subscribe 订阅主题
	Subscribe(topic string, selector MessageSelector, listener MessageListener) error

	// Unsubscribe 取消订阅
	Unsubscribe(topic string) error

	// Start 启动消费者
	Start() error

	// Shutdown 关闭消费者
	Shutdown() error
}

// MessageSelector 消息选择器（用于过滤消息）
type MessageSelector struct {
	// Type 选择器类型：TAG 或 SQL92
	Type string

	// Expression 表达式
	// 如果是 TAG 类型，表达式为标签，如 "tag1 || tag2"
	// 如果是 SQL92 类型，表达式为 SQL 条件，如 "a > 10 AND b = 'test'"
	Expression string
}

// NewTagSelector 创建标签选择器
func NewTagSelector(tags string) MessageSelector {
	return MessageSelector{
		Type:       "TAG",
		Expression: tags,
	}
}

// NewSQLSelector 创建 SQL 选择器
func NewSQLSelector(expression string) MessageSelector {
	return MessageSelector{
		Type:       "SQL92",
		Expression: expression,
	}
}

// DefaultConsumer 默认消费者实现
// 注意：这是一个抽象实现，需要根据实际使用的 RocketMQ 客户端库进行适配
type DefaultConsumer struct {
	config     Config
	consumer   rmq.PushConsumer
	subscribes map[string]subscription
	mu         sync.RWMutex
	started    bool
	logger     logx.Logger
}

type subscription struct {
	selector MessageSelector
	listener MessageListener
}

// NewConsumer 创建消费者
func NewConsumer(config Config) (Consumer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	if err := config.Consumer.Validate(); err != nil {
		return nil, fmt.Errorf("消费者配置验证失败: %w", err)
	}

	return &DefaultConsumer{
		config:     config,
		subscribes: make(map[string]subscription),
		logger:     logx.WithContext(context.Background()),
	}, nil
}

// Subscribe 订阅主题
func (c *DefaultConsumer) Subscribe(topic string, selector MessageSelector, listener MessageListener) error {
	if topic == "" {
		return fmt.Errorf("主题不能为空")
	}
	if listener == nil {
		return fmt.Errorf("消息监听器不能为空")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("消费者已启动，暂不支持动态订阅")
	}

	c.subscribes[topic] = subscription{selector: selector, listener: listener}
	c.logger.Infof("订阅主题: Topic=%s, Selector=%s", topic, selector.Expression)

	return nil
}

// Unsubscribe 取消订阅
func (c *DefaultConsumer) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.subscribes, topic)
	c.logger.Infof("取消订阅主题: Topic=%s", topic)

	return nil
}

// Start 启动消费者
func (c *DefaultConsumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("消费者已经启动")
	}

	if len(c.subscribes) == 0 {
		return fmt.Errorf("没有订阅任何主题，请先调用 Subscribe()")
	}

	opts := []consumer.Option{
		consumer.WithNameServer(semicolonSplit(c.config.NameServer)),
		consumer.WithGroupName(c.config.Consumer.GroupName),
		consumer.WithConsumerModel(toConsumeModel(c.config.Consumer.MessageModel)),
		consumer.WithConsumeMessageBatchMaxSize(c.config.Consumer.ConsumeMessageBatchMaxSize),
		consumer.WithMaxReconsumeTimes(int32(c.config.Consumer.MaxReconsumeTimes)),
	}
	cons, err := rmq.NewPushConsumer(opts...)
	if err != nil {
		return fmt.Errorf("创建消费者失败: %w", err)
	}

	for topic, sub := range c.subscribes {
		sel := toSelector(sub.selector)
		err := cons.Subscribe(topic, sel, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			messages := make([]*Message, 0, len(msgs))
			for _, m := range msgs {
				if m == nil {
					continue
				}
				messages = append(messages, &Message{
					Topic:      m.Topic,
					Tag:        m.GetTags(),
					Key:        m.GetKeys(),
					Body:       m.Body,
					Properties: m.GetProperties(),
					MessageID:  m.MsgId,
				})
			}

			status, err := sub.listener.ConsumeMessage(ctx, messages)
			if err != nil {
				c.logger.Errorf("消费消息失败: %v", err)
				return consumer.ConsumeRetryLater, err
			}
			if status == ConsumeLater {
				return consumer.ConsumeRetryLater, nil
			}
			return consumer.ConsumeSuccess, nil
		})
		if err != nil {
			return fmt.Errorf("订阅主题失败: %w", err)
		}
	}

	if err := cons.Start(); err != nil {
		return fmt.Errorf("启动消费者失败: %w", err)
	}
	c.consumer = cons

	c.started = true
	c.logger.Infof("RocketMQ 消费者启动成功, NameServer: %s, Group: %s, 订阅主题数: %d",
		c.config.NameServer, c.config.Consumer.GroupName, len(c.subscribes))

	return nil
}

// Shutdown 关闭消费者
func (c *DefaultConsumer) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	if c.consumer != nil {
		if err := c.consumer.Shutdown(); err != nil {
			return fmt.Errorf("关闭消费者失败: %w", err)
		}
	}

	c.started = false
	c.subscribes = make(map[string]subscription)
	c.logger.Infof("RocketMQ 消费者已关闭")

	return nil
}

// SimpleMessageListener 简单消息监听器（用于快速实现）
type SimpleMessageListener struct {
	Handler func(ctx context.Context, msgs []*Message) error
}

// ConsumeMessage 消费消息
func (l *SimpleMessageListener) ConsumeMessage(ctx context.Context, msgs []*Message) (ConsumeStatus, error) {
	if l.Handler == nil {
		return ConsumeSuccess, nil
	}

	err := l.Handler(ctx, msgs)
	if err != nil {
		return ConsumeLater, err
	}

	return ConsumeSuccess, nil
}

func toConsumeModel(model string) consumer.MessageModel {
	switch strings.ToUpper(model) {
	case "BROADCASTING":
		return consumer.BroadCasting
	default:
		return consumer.Clustering
	}
}

func toSelector(sel MessageSelector) consumer.MessageSelector {
	t := strings.ToUpper(sel.Type)
	expr := sel.Expression
	if expr == "" {
		expr = "*"
	}
	switch t {
	case "SQL92":
		return consumer.MessageSelector{
			Type:       consumer.SQL92,
			Expression: expr,
		}
	default:
		return consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: expr,
		}
	}
}
