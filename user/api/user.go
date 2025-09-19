package main

import (
	"flag"
	"fmt"

	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/user/api/internal/config"
	"github.com/hd2yao/ecshop/user/api/internal/handler"
	"github.com/hd2yao/ecshop/user/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	
	// 根据配置使用Gin风格的日志中间件
	if c.LogFormat == "gin" {
		server.Use(middleware.GinStyleLogger())
	}

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
