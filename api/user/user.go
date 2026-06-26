package main

import (
	"flag"
	"fmt"

	"gozero-demo/api/user/internal/config"
	"gozero-demo/api/user/internal/handler"
	"gozero-demo/api/user/internal/middleware"
	"gozero-demo/api/user/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var (
	configFile = flag.String("f", "etc/user-api.yaml", "the config file")
	version    = "dev"
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 注册全局中间件
	// 执行顺序：内置中间件 → JWT(per-route) → Logging → RateLimit → Handler
	server.Use(middleware.LoggingMiddleware)
	if ctx.RateLimit != nil {
		server.Use(ctx.RateLimit.Handle)
	}

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting user-api %s at %s:%d...\n", version, c.Host, c.Port)
	server.Start()
}
