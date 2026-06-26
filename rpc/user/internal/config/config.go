package config

import (
	"gozero-demo/internal/database"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB database.DBConf // GORM 数据库配置
}
