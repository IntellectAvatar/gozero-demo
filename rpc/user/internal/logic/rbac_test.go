package logic

import (
	"context"
	"testing"

	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRBACTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.Role{}, &model.Permission{}, &model.UserRole{}, &model.RolePermission{})

	// 种子数据
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pwd"), bcrypt.DefaultCost)
	db.Create(&model.User{Username: "admin", Password: string(hashed), Email: "admin@test.com"})
	db.Create(&model.Role{Code: "admin", Name: "管理员"})
	db.Create(&model.Role{Code: "user", Name: "用户"})
	db.Create(&model.Permission{Code: "user:read", Name: "查用户"})
	db.Create(&model.Permission{Code: "user:write", Name: "改用户"})
	return db
}

func TestAssignRole(t *testing.T) {
	db := setupRBACTestDB(t)
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewAssignRoleLogic(context.Background(), svcCtx)
	_, err := l.AssignRole(&pb.AssignRoleRequest{UserId: 1, RoleId: 1})
	if err != nil {
		t.Fatalf("分配角色失败: %v", err)
	}

	// 验证
	var ur model.UserRole
	db.Where("user_id = ? AND role_id = ?", 1, 1).First(&ur)
	if ur.UserID != 1 {
		t.Error("user_role 未插入")
	}
}

func TestGetUserRoles_WithPermissions(t *testing.T) {
	db := setupRBACTestDB(t)
	db.Create(&model.UserRole{UserID: 1, RoleID: 1})
	db.Create(&model.RolePermission{RoleID: 1, PermissionID: 1})
	db.Create(&model.RolePermission{RoleID: 1, PermissionID: 2})
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewGetUserRolesLogic(context.Background(), svcCtx)
	resp, err := l.GetUserRoles(&pb.GetUserRolesRequest{UserId: 1})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(resp.Roles) != 1 {
		t.Errorf("Roles = %d, want 1", len(resp.Roles))
	}
	if len(resp.Permissions) != 2 {
		t.Errorf("Permissions = %d, want 2", len(resp.Permissions))
	}
}

func TestRemoveRole(t *testing.T) {
	db := setupRBACTestDB(t)
	db.Create(&model.UserRole{UserID: 1, RoleID: 1})
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewRemoveRoleLogic(context.Background(), svcCtx)
	_, err := l.RemoveRole(&pb.RemoveRoleRequest{UserId: 1, RoleId: 1})
	if err != nil {
		t.Fatalf("移除失败: %v", err)
	}

	var count int64
	db.Model(&model.UserRole{}).Where("user_id = ?", 1).Count(&count)
	if count != 0 {
		t.Error("角色未移除")
	}
}

func TestListRoles(t *testing.T) {
	db := setupRBACTestDB(t)
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewListRolesLogic(context.Background(), svcCtx)
	resp, err := l.ListRoles(&pb.ListRolesRequest{})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(resp.Roles) != 2 {
		t.Errorf("Roles = %d, want 2 (admin + user)", len(resp.Roles))
	}
}

func TestListPermissions(t *testing.T) {
	db := setupRBACTestDB(t)
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewListPermissionsLogic(context.Background(), svcCtx)
	resp, err := l.ListPermissions(&pb.ListPermissionsRequest{})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if len(resp.Permissions) != 2 {
		t.Errorf("Permissions = %d, want 2", len(resp.Permissions))
	}
}

func TestAssignRolePermission(t *testing.T) {
	db := setupRBACTestDB(t)
	db.Create(&model.Role{Code: "editor", Name: "编辑"})
	db.Create(&model.Permission{Code: "article:write", Name: "写文章"})
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewAssignRolePermissionLogic(context.Background(), svcCtx)
	_, err := l.AssignRolePermission(&pb.AssignRolePermissionRequest{RoleId: 3, PermissionId: 3})
	if err != nil {
		t.Fatalf("分配权限失败: %v", err)
	}

	var rp model.RolePermission
	db.Where("role_id = ? AND permission_id = ?", 3, 3).First(&rp)
	if rp.RoleID != 3 {
		t.Error("role_permission 未插入")
	}
}

func TestRemoveRolePermission(t *testing.T) {
	db := setupRBACTestDB(t)
	db.Create(&model.RolePermission{RoleID: 1, PermissionID: 1})
	svcCtx := &svc.ServiceContext{DB: db}

	l := NewRemoveRolePermissionLogic(context.Background(), svcCtx)
	_, err := l.RemoveRolePermission(&pb.RemoveRolePermissionRequest{RoleId: 1, PermissionId: 1})
	if err != nil {
		t.Fatalf("移除权限失败: %v", err)
	}

	var count int64
	db.Model(&model.RolePermission{}).Where("role_id = ?", 1).Count(&count)
	if count != 0 {
		t.Error("权限未移除")
	}
}
