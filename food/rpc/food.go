package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hd2yao/ecshop/common/rocketmq"
	"github.com/hd2yao/ecshop/food/model"
	"github.com/hd2yao/ecshop/food/rpc/internal/config"
	"github.com/hd2yao/ecshop/food/rpc/internal/server"
	"github.com/hd2yao/ecshop/food/rpc/internal/svc"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

var configFile = flag.String("f", "etc/food.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// 启动 RocketMQ 消费者（如果配置了）
	if c.RocketMQ.NameServer != "" {
		fmt.Printf("正在初始化 RocketMQ, NameServer: %s\n", c.RocketMQ.NameServer)
		// 初始化 RocketMQ
		if err := rocketmq.InitRocketMQ(c.RocketMQ); err != nil {
			panic(fmt.Sprintf("初始化 RocketMQ 失败: %v", err))
		}
		fmt.Println("RocketMQ 初始化成功")
		
		// 先启动生产者并发送一条消息来创建 Topic（如果不存在）
		if err := ensureTopicExists(c.RocketMQ); err != nil {
			fmt.Printf("警告: 无法确保 Topic 存在: %v\n", err)
		}
		
		startRocketMQConsumer(ctx)
	} else {
		fmt.Println("RocketMQ 未配置，跳过消费者启动")
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		food.RegisterFoodServer(grpcServer, server.NewFoodServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// startRocketMQConsumer 启动 RocketMQ 消费者
func startRocketMQConsumer(ctx *svc.ServiceContext) {
	fmt.Println("正在启动 RocketMQ 消费者...")
	
	rmq := rocketmq.GetRocketMQ()

	// 创建美食更新监听器
	listener := model.NewFoodUpdateListener(ctx.FoodModel, ctx.FoodCache)

	// 订阅主题（使用标签过滤：create || update || delete）
	selector := rocketmq.NewTagSelector(fmt.Sprintf("%s || %s || %s", model.FoodTagCreate, model.FoodTagUpdate, model.FoodTagDelete))
	fmt.Printf("正在订阅主题: %s\n", model.MQTopicFoodUpdate)
	if err := rmq.Subscribe(model.MQTopicFoodUpdate, selector, listener); err != nil {
		panic(fmt.Sprintf("订阅 RocketMQ 主题失败: %v", err))
	}

	// 启动消费者
	fmt.Println("正在启动消费者...")
	if err := rmq.StartConsumer(); err != nil {
		panic(fmt.Sprintf("启动 RocketMQ 消费者失败: %v", err))
	}

	fmt.Printf("RocketMQ 消费者启动成功，已订阅主题: %s\n", model.MQTopicFoodUpdate)
}

// ensureTopicExists 确保 Topic 存在（通过发送一条消息来触发自动创建）
func ensureTopicExists(config rocketmq.Config) error {
	rmq := rocketmq.GetRocketMQ()
	
	// 获取生产者（会触发创建和启动）
	_, err := rmq.GetProducer()
	if err != nil {
		return fmt.Errorf("获取生产者失败: %w", err)
	}
	
	// 发送一条空消息来触发 Topic 的自动创建
	// 使用一个特殊的初始化消息
	initMsg := map[string]interface{}{
		"type": "init",
		"msg":  "topic_initialization",
	}
	
	// 异步发送，不等待结果（因为这只是为了创建 Topic）
	_ = rmq.SendMessageAsync(context.Background(), model.MQTopicFoodUpdate, model.FoodTagInit, initMsg, func(result *rocketmq.SendResult, err error) {
		if err != nil {
			fmt.Printf("Topic 初始化消息发送失败（可忽略）: %v\n", err)
		} else {
			fmt.Println("Topic 初始化消息发送成功，Topic 已创建")
		}
	})
	
	// 等待一小段时间，让消息发送完成
	time.Sleep(500 * time.Millisecond)
	
	return nil
}
