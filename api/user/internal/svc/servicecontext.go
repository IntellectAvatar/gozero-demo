package svc

import (
	"gozero-demo/api/user/internal/config"
	"gozero-demo/api/user/internal/middleware"
	"gozero-demo/internal/database"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config    config.Config
	RateLimit *middleware.RateLimiter
	UserRpc   pb.UserClient
	DB        *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	svc := &ServiceContext{Config: c}

	// 限流器
	if c.RateLimit.Enabled {
		svc.RateLimit = middleware.NewRateLimiter(c.RateLimit.RPS, c.RateLimit.Burst)
	}

	// user RPC 客户端（etcd 服务发现）
	conn, err := zrpc.NewClient(c.UserRpc)
	if err != nil {
		logx.Errorf("user RPC 客户端初始化失败（etcd 可能未启动），将启用降级模式: %v", err)
	} else {
		svc.UserRpc = pb.NewUserClient(conn.Conn())
		logx.Infof("user RPC 客户端已初始化")
	}

	// GORM 数据库连接
	svc.DB = database.MustInitGorm(c.DB)

	return svc
}
