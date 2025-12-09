package rocketmq

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	rmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/zeromicro/go-zero/core/logx"
)

// Producer RocketMQ 生产者接口
type Producer interface {
	// Send 发送消息
	Send(ctx context.Context, msg *Message) (*SendResult, error)

	// SendAsync 异步发送消息
	SendAsync(ctx context.Context, msg *Message, callback func(*SendResult, error)) error

	// SendBatch 批量发送消息
	SendBatch(ctx context.Context, msgs []*Message) (*SendResult, error)

	// Start 启动生产者
	Start() error

	// Shutdown 关闭生产者
	Shutdown() error
}

// SendResult 发送结果
type SendResult struct {
	// MessageID 消息ID
	MessageID string `json:"message_id"`

	// QueueID 队列ID
	QueueID int `json:"queue_id"`

	// QueueOffset 队列偏移量
	QueueOffset int64 `json:"queue_offset"`

	// TransactionID 事务ID（事务消息）
	TransactionID string `json:"transaction_id,omitempty"`

	// RegionID 区域ID
	RegionID string `json:"region_id,omitempty"`

	// TraceOn 是否开启消息轨迹
	TraceOn bool `json:"trace_on,omitempty"`
}

// DefaultProducer 默认生产者实现
type DefaultProducer struct {
	config   Config
	producer rmq.Producer
	mu       sync.RWMutex
	started  bool
	logger   logx.Logger
}

// NewProducer 创建生产者
func NewProducer(config Config) (Producer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	if err := config.Producer.Validate(); err != nil {
		return nil, fmt.Errorf("生产者配置验证失败: %w", err)
	}

	return &DefaultProducer{
		config: config,
		logger: logx.WithContext(context.Background()),
	}, nil
}

// Start 启动生产者
func (p *DefaultProducer) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("生产者已经启动")
	}

	nameServers := semicolonSplit(p.config.NameServer)
	// 解析主机名为 IP 地址
	resolvedServers := make([]string, 0, len(nameServers))
	for _, addr := range nameServers {
		resolved := resolveNameServerAddress(addr)
		resolvedServers = append(resolvedServers, resolved)
		if resolved != addr {
			p.logger.Infof("NameServer 地址解析: %s -> %s", addr, resolved)
		}
	}
	p.logger.Infof("RocketMQ NameServer 地址: %v (解析后: %v)", nameServers, resolvedServers)

	opts := []producer.Option{
		producer.WithNameServer(resolvedServers),
		producer.WithGroupName(p.config.Producer.GroupName),
		producer.WithRetry(p.config.Producer.RetryTimes),
		producer.WithSendMsgTimeout(time.Duration(p.config.Producer.SendMsgTimeout) * time.Millisecond),
	}
	if p.config.Producer.CompressLevel > 0 {
		opts = append(opts, producer.WithCompressLevel(p.config.Producer.CompressLevel))
	}
	if p.config.Producer.VipChannelEnabled {
		opts = append(opts, producer.WithVIPChannel(true))
	}

	prod, err := rmq.NewProducer(opts...)
	if err != nil {
		return fmt.Errorf("创建生产者失败: %w", err)
	}
	if err := prod.Start(); err != nil {
		return fmt.Errorf("启动生产者失败: %w", err)
	}
	p.producer = prod

	p.started = true
	p.logger.Infof("RocketMQ 生产者启动成功, NameServer: %s, Group: %s",
		p.config.NameServer, p.config.Producer.GroupName)

	return nil
}

// Send 发送消息
func (p *DefaultProducer) Send(ctx context.Context, msg *Message) (*SendResult, error) {
	p.mu.RLock()
	if !p.started {
		p.mu.RUnlock()
		return nil, fmt.Errorf("生产者未启动，请先调用 Start()")
	}
	p.mu.RUnlock()

	if err := p.validateMessage(msg); err != nil {
		return nil, err
	}

	rmqMsg := primitive.NewMessage(msg.Topic, msg.Body)
	if msg.Tag != "" {
		rmqMsg.WithTag(msg.Tag)
	}
	if msg.Key != "" {
		rmqMsg.WithKeys([]string{msg.Key})
	}
	if msg.DelayTimeLevel > 0 {
		rmqMsg.WithDelayTimeLevel(msg.DelayTimeLevel)
	}
	for k, v := range msg.Properties {
		rmqMsg.WithProperty(k, v)
	}

	result, err := p.producer.SendSync(ctx, rmqMsg)
	if err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}

	return &SendResult{
		MessageID:   result.MsgID,
		QueueID:     int(result.MessageQueue.QueueId),
		QueueOffset: result.QueueOffset,
		RegionID:    result.RegionID,
		TraceOn:     result.TraceOn,
	}, nil
}

// SendAsync 异步发送消息
func (p *DefaultProducer) SendAsync(ctx context.Context, msg *Message, callback func(*SendResult, error)) error {
	p.mu.RLock()
	if !p.started {
		p.mu.RUnlock()
		return fmt.Errorf("生产者未启动，请先调用 Start()")
	}
	p.mu.RUnlock()

	if err := p.validateMessage(msg); err != nil {
		if callback != nil {
			callback(nil, err)
		}
		return err
	}

	rmqMsg := primitive.NewMessage(msg.Topic, msg.Body)
	if msg.Tag != "" {
		rmqMsg.WithTag(msg.Tag)
	}
	if msg.Key != "" {
		rmqMsg.WithKeys([]string{msg.Key})
	}
	if msg.DelayTimeLevel > 0 {
		rmqMsg.WithDelayTimeLevel(msg.DelayTimeLevel)
	}
	for k, v := range msg.Properties {
		rmqMsg.WithProperty(k, v)
	}

	return p.producer.SendAsync(ctx, func(ctx context.Context, result *primitive.SendResult, err error) {
		if callback == nil {
			return
		}
		if err != nil {
			callback(nil, err)
			return
		}
		callback(&SendResult{
			MessageID:   result.MsgID,
			QueueID:     int(result.MessageQueue.QueueId),
			QueueOffset: result.QueueOffset,
			RegionID:    result.RegionID,
			TraceOn:     result.TraceOn,
		}, nil)
	}, rmqMsg)
}

// SendBatch 批量发送消息
func (p *DefaultProducer) SendBatch(ctx context.Context, msgs []*Message) (*SendResult, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("消息列表为空")
	}

	// 验证所有消息
	for i, msg := range msgs {
		if err := p.validateMessage(msg); err != nil {
			return nil, fmt.Errorf("第 %d 条消息验证失败: %w", i+1, err)
		}
	}

	// TODO: 根据实际使用的 RocketMQ 客户端库批量发送消息
	// 注意：RocketMQ 批量消息要求所有消息的 Topic 相同

	// 临时实现：逐个发送
	var lastResult *SendResult
	for _, msg := range msgs {
		result, err := p.Send(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("批量发送消息失败: %w", err)
		}
		lastResult = result
	}

	return lastResult, nil
}

// Shutdown 关闭生产者
func (p *DefaultProducer) Shutdown() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return nil
	}

	if p.producer != nil {
		if err := p.producer.Shutdown(); err != nil {
			return fmt.Errorf("关闭生产者失败: %w", err)
		}
	}

	p.started = false
	p.logger.Infof("RocketMQ 生产者已关闭")
	return nil
}

// validateMessage 验证消息
func (p *DefaultProducer) validateMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("消息不能为空")
	}
	if msg.Topic == "" {
		return fmt.Errorf("消息 Topic 不能为空")
	}
	if len(msg.Body) == 0 {
		return fmt.Errorf("消息体不能为空")
	}
	if len(msg.Body) > p.config.Producer.MaxMessageSize {
		return fmt.Errorf("消息体大小超过限制: %d > %d", len(msg.Body), p.config.Producer.MaxMessageSize)
	}
	return nil
}

func semicolonSplit(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ';' || r == ','
	})
	return parts
}
