package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hd2yao/ecshop/common/rocketmq"
	"github.com/hd2yao/ecshop/social/rpc/internal/config"
	"github.com/hd2yao/ecshop/social/rpc/internal/logic"
	"github.com/hd2yao/ecshop/social/rpc/internal/server"
	"github.com/hd2yao/ecshop/social/rpc/internal/svc"
	"github.com/hd2yao/ecshop/social/rpc/types/social"
)

var configFile = flag.String("f", "etc/social.yaml", "the config file")

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

		startRocketMQConsumer(ctx)
	} else {
		fmt.Println("RocketMQ 未配置，跳过消费者启动")
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		social.RegisterSocialServer(grpcServer, server.NewSocialServer(ctx))

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

	// 创建关注/取关事件监听器
	listener := logic.NewFollowEventListener(ctx)

	// 订阅主题（使用标签过滤：blogger || follower）
	selector := rocketmq.NewTagSelector("blogger || follower")
	fmt.Println("正在订阅主题: social_follow")
	if err := rmq.Subscribe(logic.MQTopicFollow, selector, listener); err != nil {
		panic(fmt.Sprintf("订阅 RocketMQ 主题失败: %v", err))
	}

	// 启动消费者
	fmt.Println("正在启动消费者...")
	if err := rmq.StartConsumer(); err != nil {
		panic(fmt.Sprintf("启动 RocketMQ 消费者失败: %v", err))
	}

	fmt.Println("RocketMQ 消费者启动成功，已订阅主题: social_follow")
}
