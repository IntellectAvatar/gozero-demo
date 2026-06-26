package config

import (
	"gozero-demo/internal/database"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// AuthConf JWT 鉴权配置
type AuthConf struct {
	AccessSecret string `json:",default=your-jwt-secret-key"`
	AccessExpire int64  `json:",default=3600"`
}

// RateLimitConf 限流配置
type RateLimitConf struct {
	Enabled bool `json:",default=false"`
	RPS     int  `json:",default=10"`
	Burst   int  `json:",default=20"`
}

type Config struct {
	rest.RestConf
	Auth      AuthConf
	RateLimit RateLimitConf
	UserRpc   zrpc.RpcClientConf
	DB        database.DBConf
}
