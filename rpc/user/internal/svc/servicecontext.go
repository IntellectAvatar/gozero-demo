package svc

import (
	"gozero-demo/internal/database"
	"gozero-demo/rpc/user/internal/config"
	"gozero-demo/rpc/user/internal/model"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := database.MustInitGorm(c.DB)

	// 自动建表
	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
	); err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
	}
}
