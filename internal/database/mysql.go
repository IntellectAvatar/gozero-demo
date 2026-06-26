package database

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// gormLogWriter 将 GORM 日志桥接到 go-zero 的 logx
type gormLogWriter struct{}

func (w gormLogWriter) Printf(format string, args ...interface{}) {
	logx.Info(fmt.Sprintf(format, args...))
}

// MustInitGorm 初始化 GORM 数据库连接，失败时 panic。
// 配置了连接池参数和慢查询日志。
func MustInitGorm(conf DBConf) *gorm.DB {
	// GORM 日志适配到 go-zero logx
	gormLogger := logger.New(
		gormLogWriter{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
			LogLevel:                  logger.Warn,            // 只记录慢查询和错误
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(conf.DSN()), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // 跳过默认事务以提升性能
	})
	if err != nil {
		logx.Errorf("数据库连接失败: %v", err)
		panic(err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		logx.Errorf("获取 sql.DB 失败: %v", err)
		panic(err)
	}
	sqlDB.SetMaxOpenConns(conf.MaxConns)
	sqlDB.SetMaxIdleConns(conf.MaxIdle)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		logx.Errorf("数据库 Ping 失败: %v", err)
		panic(err)
	}

	logx.Infof("数据库连接成功: %s:%d/%s", conf.Host, conf.Port, conf.Database)
	return db
}
