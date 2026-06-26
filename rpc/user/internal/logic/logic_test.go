package logic

import (
	"context"
	"fmt"
	"testing"

	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRpcTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("连接 SQLite 失败: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("test-pass"), bcrypt.DefaultCost)
	db.Create(&model.User{Username: "testuser", Password: string(hashed), Email: "test@test.com", Phone: "13800000001"})
	return db
}

func newRpcSvcCtx(db *gorm.DB) *svc.ServiceContext {
	return &svc.ServiceContext{DB: db}
}

// ====================== GetUser ======================

func TestGetUser_Exists(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewGetUserLogic(context.Background(), svcCtx)
	resp, err := l.GetUser(&pb.GetUserRequest{Id: 1})
	if err != nil {
		t.Fatalf("GetUser 失败: %v", err)
	}
	if resp.User.Username != "testuser" {
		t.Errorf("Username = %s, want testuser", resp.User.Username)
	}
	if resp.User.Email != "test@test.com" {
		t.Errorf("Email = %s, want test@test.com", resp.User.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewGetUserLogic(context.Background(), svcCtx)
	_, err := l.GetUser(&pb.GetUserRequest{Id: 999})
	if err == nil {
		t.Fatal("不存在的用户应返回错误")
	}
}

// ====================== CreateUser ======================

func TestCreateUser_Success(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewCreateUserLogic(context.Background(), svcCtx)
	resp, err := l.CreateUser(&pb.CreateUserRequest{
		Username: "newuser",
		Password: "new-pass",
		Email:    "new@test.com",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if resp.User.Id == 0 {
		t.Error("用户 ID 不应为 0")
	}
	if resp.User.Username != "newuser" {
		t.Errorf("Username = %s, want newuser", resp.User.Username)
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewCreateUserLogic(context.Background(), svcCtx)
	_, err := l.CreateUser(&pb.CreateUserRequest{
		Username: "testuser", // 已存在
		Password: "any",
		Email:    "any@test.com",
	})
	if err == nil {
		t.Fatal("重复用户名应返回错误")
	}
}

// ====================== UpdateUser ======================

func TestUpdateUser_Success(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewUpdateUserLogic(context.Background(), svcCtx)
	resp, err := l.UpdateUser(&pb.UpdateUserRequest{Id: 1, Email: "updated@test.com"})
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}
	if resp.User.Email != "updated@test.com" {
		t.Errorf("Email = %s, want updated@test.com", resp.User.Email)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewUpdateUserLogic(context.Background(), svcCtx)
	_, err := l.UpdateUser(&pb.UpdateUserRequest{Id: 999, Email: "x"})
	if err == nil {
		t.Fatal("不存在用户应返回错误")
	}
}

// ====================== UpdatePassword ======================

func TestUpdatePassword_CorrectOldPassword(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewUpdatePasswordLogic(context.Background(), svcCtx)
	_, err := l.UpdatePassword(&pb.UpdatePasswordRequest{
		Id:          1,
		OldPassword: "test-pass",
		NewPassword: "new-secret",
	})
	if err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}
}

func TestUpdatePassword_WrongOldPassword(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewUpdatePasswordLogic(context.Background(), svcCtx)
	_, err := l.UpdatePassword(&pb.UpdatePasswordRequest{
		Id:          1,
		OldPassword: "wrong-pass",
		NewPassword: "new-secret",
	})
	if err == nil {
		t.Fatal("旧密码错误应返回错误")
	}
}

// ====================== ListUsers ======================

func TestListUsers_Pagination(t *testing.T) {
	db := setupRpcTestDB(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pwd"), bcrypt.DefaultCost)
	for i := 0; i < 10; i++ {
		db.Create(&model.User{
			Username: "extra-" + string(rune('a'+i)),
			Password: string(hashed),
			Email:    "test@test.com",
			Phone:    fmt.Sprintf("1380000000%d", i+2),
		})
	}
	svcCtx := newRpcSvcCtx(db)

	l := NewListUsersLogic(context.Background(), svcCtx)
	resp, err := l.ListUsers(&pb.ListUsersRequest{Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("列表查询失败: %v", err)
	}
	if len(resp.Users) != 5 {
		t.Errorf("len(Users) = %d, want 5", len(resp.Users))
	}
	if resp.Total != 11 {
		t.Errorf("Total = %d, want 11", resp.Total)
	}
}

func TestListUsers_EmptyPage(t *testing.T) {
	db := setupRpcTestDB(t)
	svcCtx := newRpcSvcCtx(db)

	l := NewListUsersLogic(context.Background(), svcCtx)
	// Page = 100 应返回空列表
	resp, err := l.ListUsers(&pb.ListUsersRequest{Page: 100, PageSize: 5})
	if err != nil {
		t.Fatalf("空页查询失败: %v", err)
	}
	if len(resp.Users) != 0 {
		t.Errorf("len(Users) = %d, want 0", len(resp.Users))
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
}
