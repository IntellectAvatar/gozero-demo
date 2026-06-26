package model

import "time"

// Role 角色表
type Role struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"Id"`
	Name        string    `gorm:"uniqueIndex;size:64;not null" json:"Name"`
	Code        string    `gorm:"uniqueIndex;size:64;not null" json:"Code"`
	Description string    `gorm:"size:256" json:"Description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"CreatedAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"UpdatedAt"`
}

func (Role) TableName() string { return "roles" }

// Permission 权限表
type Permission struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"Id"`
	Code        string    `gorm:"uniqueIndex;size:64;not null" json:"Code"`
	Name        string    `gorm:"size:64;not null" json:"Name"`
	Description string    `gorm:"size:256" json:"Description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"CreatedAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"UpdatedAt"`
}

func (Permission) TableName() string { return "permissions" }

// UserRole 用户-角色关联
type UserRole struct {
	UserID int64 `gorm:"primaryKey" json:"UserId"`
	RoleID int64 `gorm:"primaryKey" json:"RoleId"`
}

func (UserRole) TableName() string { return "user_roles" }

// RolePermission 角色-权限关联
type RolePermission struct {
	RoleID       int64 `gorm:"primaryKey" json:"RoleId"`
	PermissionID int64 `gorm:"primaryKey" json:"PermissionId"`
}

func (RolePermission) TableName() string { return "role_permissions" }
