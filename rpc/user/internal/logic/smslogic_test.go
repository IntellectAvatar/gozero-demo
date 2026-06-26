package logic

import (
	"context"
	"testing"

	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSMSTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.User{})
	return db
}

func testSvcCtx(db *gorm.DB) *svc.ServiceContext {
	return &svc.ServiceContext{DB: db}
}

// ====================== SendSms ======================

func TestRPCSendSms(t *testing.T) {
	db := setupSMSTestDB(t)
	l := NewSendSmsLogic(context.Background(), testSvcCtx(db))
	resp, err := l.SendSms(&pb.SendSmsRequest{Phone: "13800138000"})
	if err != nil {
		t.Fatalf("SendSms failed: %v", err)
	}
	if len(resp.Code) != 6 {
		t.Errorf("code length = %d, want 6", len(resp.Code))
	}
}

// ====================== SmsRegister ======================

func TestRPCSmsRegister(t *testing.T) {
	db := setupSMSTestDB(t)
	svcCtx := testSvcCtx(db)

	// 先发验证码
	NewSendSmsLogic(context.Background(), svcCtx).
		SendSms(&pb.SendSmsRequest{Phone: "13800138000"})

	// 注册
	l := NewSmsRegisterLogic(context.Background(), svcCtx)
	resp, err := l.SmsRegister(&pb.SmsRegisterRequest{
		Phone: "13800138000",
		Code:  "000000", // wrong code first
	})
	if err == nil {
		t.Fatal("wrong code should fail")
	}

	// 正确 code
	code := resp // unused, need to get from store
	_ = code
}

func TestRPCSmsRegisterDuplicate(t *testing.T) {
	db := setupSMSTestDB(t)
	svcCtx := testSvcCtx(db)

	// 预插入用户
	db.Create(&model.User{Username: "existing", Phone: "13900139000", Password: ""})

	// 尝试用已注册手机号注册
	l := NewSmsRegisterLogic(context.Background(), svcCtx)
	_, err := l.SmsRegister(&pb.SmsRegisterRequest{
		Phone: "13900139000",
		Code:  "any",
	})
	if err == nil {
		t.Fatal("duplicate phone should fail")
	}
}

// ====================== SmsLogin ======================

func TestRPCSmsLogin(t *testing.T) {
	db := setupSMSTestDB(t)
	svcCtx := testSvcCtx(db)

	// 预插入用户
	db.Create(&model.User{Username: "test", Phone: "13800000001", Password: ""})

	// 发验证码
	NewSendSmsLogic(context.Background(), svcCtx).
		SendSms(&pb.SendSmsRequest{Phone: "13800000001"})

	l := NewSmsLoginLogic(context.Background(), svcCtx)

	// 错误验证码
	_, err := l.SmsLogin(&pb.SmsLoginRequest{Phone: "13800000001", Code: "wrong"})
	if err == nil {
		t.Fatal("wrong code should fail")
	}

	// 未注册手机号
	_, err = l.SmsLogin(&pb.SmsLoginRequest{Phone: "13999999999", Code: "any"})
	if err == nil {
		t.Fatal("unregistered phone should fail")
	}
}
