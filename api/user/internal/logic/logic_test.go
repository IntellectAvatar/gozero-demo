package logic

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gozero-demo/api/user/internal/config"
	"gozero-demo/api/user/internal/middleware"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/rpc/user/pb"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建内存 SQLite + 建表 + 插入测试用户
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	// RBAC 表
	db.Exec(`CREATE TABLE IF NOT EXISTS roles (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT UNIQUE, name TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS permissions (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT UNIQUE, name TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS user_roles (user_id INTEGER, role_id INTEGER, PRIMARY KEY (user_id, role_id))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS role_permissions (role_id INTEGER, permission_id INTEGER, PRIMARY KEY (role_id, permission_id))`)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("test-pass"), bcrypt.DefaultCost)
	db.Exec(`INSERT INTO users (username, password, email) VALUES (?, ?, ?)`,
		"testuser", string(hashed), "test@test.com")
	return db
}

func newTestSvcCtx(db *gorm.DB) *svc.ServiceContext {
	return &svc.ServiceContext{
		Config: config.Config{
			Auth: config.AuthConf{AccessSecret: "test-secret", AccessExpire: 3600},
		},
		DB: db,
	}
}

// ====================== Login ======================

func TestLogin_CorrectPassword(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewLoginLogic(context.Background(), svcCtx)
	resp, err := l.Login(&types.LoginRequest{
		Username: "testuser",
		Password: "test-pass",
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if resp.Token == "" {
		t.Error("Token 不应为空")
	}
	if resp.Expire == 0 {
		t.Error("Expire 不应为 0")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewLoginLogic(context.Background(), svcCtx)
	_, err := l.Login(&types.LoginRequest{
		Username: "testuser",
		Password: "wrong-pass",
	})
	if err == nil {
		t.Fatal("错误密码应返回错误")
	}
	if err.Error() != "用户名或密码错误" {
		t.Errorf("错误信息 = %s, want 用户名或密码错误", err.Error())
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewLoginLogic(context.Background(), svcCtx)
	_, err := l.Login(&types.LoginRequest{
		Username: "nonexistent",
		Password: "any",
	})
	if err == nil {
		t.Fatal("不存在的用户应返回错误")
	}
}

// ====================== UserInfo ======================

func TestUserInfo_NilRpcFallback(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)
	svcCtx.UserRpc = nil // 模拟 rpc 不可用

	ctx := context.WithValue(context.Background(), "userId", json.Number("1"))
	l := NewUserInfoLogic(ctx, svcCtx)
	resp, err := l.UserInfo()
	if err != nil {
		t.Fatalf("降级查询失败: %v", err)
	}
	if resp.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", resp.Username)
	}
	if resp.Email != "test@test.com" {
		t.Errorf("Email = %s, want test@test.com", resp.Email)
	}
}

func TestUserInfo_NoUserId(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewUserInfoLogic(context.Background(), svcCtx)
	_, err := l.UserInfo()
	if err == nil {
		t.Fatal("无 userId 时不应成功")
	}
}

// ====================== UpdateUser ======================

func TestUpdateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	ctx := context.WithValue(context.Background(), "userId", json.Number("1"))
	l := NewUpdateUserLogic(ctx, svcCtx)
	resp, err := l.UpdateUser(&types.UpdateUserRequest{Email: "new@test.com"})
	if err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if resp.Email != "new@test.com" {
		t.Errorf("Email = %s, want new@test.com", resp.Email)
	}
}

func TestUpdateUser_NoUserId(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewUpdateUserLogic(context.Background(), svcCtx)
	_, err := l.UpdateUser(&types.UpdateUserRequest{Email: "x"})
	if err == nil {
		t.Fatal("无 userId 不应成功")
	}
}

// ====================== UpdatePassword ======================

func TestUpdatePassword_CorrectOldPassword(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	ctx := context.WithValue(context.Background(), "userId", json.Number("1"))
	l := NewUpdatePasswordLogic(ctx, svcCtx)
	err := l.UpdatePassword(&types.UpdatePasswordRequest{
		OldPassword: "test-pass",
		NewPassword: "new-pass",
	})
	if err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}

	// 验证新密码可用
	loginLogic := NewLoginLogic(ctx, svcCtx)
	_, loginErr := loginLogic.Login(&types.LoginRequest{
		Username: "testuser",
		Password: "new-pass",
	})
	if loginErr != nil {
		t.Errorf("新密码登录失败: %v", loginErr)
	}
}

func TestUpdatePassword_WrongOldPassword(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	ctx := context.WithValue(context.Background(), "userId", json.Number("1"))
	l := NewUpdatePasswordLogic(ctx, svcCtx)
	err := l.UpdatePassword(&types.UpdatePasswordRequest{
		OldPassword: "wrong-old",
		NewPassword: "new-pass",
	})
	if err == nil {
		t.Fatal("旧密码错误应返回错误")
	}
}

// ====================== ListUsers ======================

func TestListUsers_Pagination(t *testing.T) {
	db := setupTestDB(t)
	// 插入额外用户
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pwd"), bcrypt.DefaultCost)
	for i := 0; i < 15; i++ {
		db.Exec("INSERT INTO users (username, password, email) VALUES (?, ?, ?)",
			"user"+string(rune('a'+i)), string(hashed), "test@test.com")
	}
	svcCtx := newTestSvcCtx(db)

	l := NewListUsersLogic(context.Background(), svcCtx)
	users, total, err := l.ListUsers(&types.ListUsersRequest{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("列表查询失败: %v", err)
	}
	if len(users) != 5 {
		t.Errorf("len(users) = %d, want 5", len(users))
	}
	if total != 16 {
		t.Errorf("total = %d, want 16 (1 initial + 15 extra)", total)
	}
}

func TestListUsers_NilRpcFallback(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)
	svcCtx.UserRpc = nil

	l := NewListUsersLogic(context.Background(), svcCtx)
	_, total, err := l.ListUsers(&types.ListUsersRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("降级列表查询失败: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
}

// ====================== 辅助 ======================

// 防止未使用 import
var _ = jwt.NewWithClaims
var _ = time.Now
var _ = middleware.GetJwtClaim
var _ = bcrypt.DefaultCost
var _ = pb.User_ServiceDesc

// ====================== RBAC ======================

func TestAssignRole_Success(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewAssignRoleLogic(context.Background(), svcCtx)
	if err := l.AssignRole(&types.AssignRoleRequest{UserID: 1, RoleID: 1}); err != nil {
		t.Fatalf("AssignRole 失败: %v", err)
	}

	var count int64
	db.Table("user_roles").Where("user_id = ? AND role_id = ?", 1, 1).Count(&count)
	if count != 1 {
		t.Error("user_roles 未插入")
	}
}

func TestRemoveRole_Success(t *testing.T) {
	db := setupTestDB(t)
	db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (1, 1)")
	svcCtx := newTestSvcCtx(db)

	l := NewRemoveRoleLogic(context.Background(), svcCtx)
	if err := l.RemoveRole(&types.RemoveRoleRequest{UserID: 1, RoleID: 1}); err != nil {
		t.Fatalf("RemoveRole 失败: %v", err)
	}

	var count int64
	db.Table("user_roles").Where("user_id = ?", 1).Count(&count)
	if count != 0 {
		t.Error("角色未移除")
	}
}

func TestListRoles_DirectDB(t *testing.T) {
	db := setupTestDB(t)
	db.Exec("INSERT INTO roles (code, name) VALUES ('admin', '管理员')")
	db.Exec("INSERT INTO roles (code, name) VALUES ('user', '用户')")
	svcCtx := newTestSvcCtx(db)

	l := NewListRolesLogic(context.Background(), svcCtx)
	roles, err := l.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles 失败: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("roles = %d, want 2", len(roles))
	}
}

func TestListPermissions_DirectDB(t *testing.T) {
	db := setupTestDB(t)
	db.Exec("INSERT INTO permissions (code, name) VALUES ('user:read', '查用户')")
	db.Exec("INSERT INTO permissions (code, name) VALUES ('user:write', '改用户')")
	svcCtx := newTestSvcCtx(db)

	l := NewListPermissionsLogic(context.Background(), svcCtx)
	perms, err := l.ListPermissions()
	if err != nil {
		t.Fatalf("ListPermissions 失败: %v", err)
	}
	if len(perms) != 2 {
		t.Errorf("perms = %d, want 2", len(perms))
	}
}

func TestAssignRolePermission_DirectDB(t *testing.T) {
	db := setupTestDB(t)
	db.Exec("INSERT INTO roles (id, code, name) VALUES (1, 'admin', '管理员')")
	db.Exec("INSERT INTO permissions (id, code, name) VALUES (1, 'user:read', '查用户')")
	svcCtx := newTestSvcCtx(db)

	l := NewAssignRolePermissionLogic(context.Background(), svcCtx)
	if err := l.AssignRolePermission(&types.AssignRolePermissionRequest{RoleID: 1, PermissionID: 1}); err != nil {
		t.Fatalf("AssignRolePermission 失败: %v", err)
	}

	var count int64
	db.Table("role_permissions").Where("role_id = ? AND permission_id = ?", 1, 1).Count(&count)
	if count != 1 {
		t.Error("role_permissions 未插入")
	}
}

func TestRemoveRolePermission_DirectDB(t *testing.T) {
	db := setupTestDB(t)
	db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (1, 1)")
	svcCtx := newTestSvcCtx(db)

	l := NewRemoveRolePermissionLogic(context.Background(), svcCtx)
	if err := l.RemoveRolePermission(&types.AssignRolePermissionRequest{RoleID: 1, PermissionID: 1}); err != nil {
		t.Fatalf("RemoveRolePermission 失败: %v", err)
	}

	var count int64
	db.Table("role_permissions").Where("role_id = ?", 1).Count(&count)
	if count != 0 {
		t.Error("权限未移除")
	}
}
