package logic

import (
	"context"
	"testing"

	"gozero-demo/api/user/internal/types"
)

func TestSmsSendApi(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	l := NewSendSmsApiLogic(context.Background(), svcCtx)
	if err := l.SendSms(&types.SendSmsRequest{Phone: "13800138000"}); err != nil {
		t.Fatalf("SendSms failed: %v", err)
	}

	// 60s 内重发应失败
	err := l.SendSms(&types.SendSmsRequest{Phone: "13800138000"})
	if err == nil {
		t.Fatal("rate limit should block second send")
	}
}

func TestSmsRegisterApi(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	// 先发验证码
	NewSendSmsApiLogic(context.Background(), svcCtx).
		SendSms(&types.SendSmsRequest{Phone: "13800138000"})

	l := NewSmsRegisterApiLogic(context.Background(), svcCtx)

	// 错误验证码
	_, err := l.SmsRegister(&types.SmsRegisterRequest{Phone: "13800138000", Code: "wrong"})
	if err == nil {
		t.Fatal("wrong code should fail")
	}
}

func TestSmsLoginApi(t *testing.T) {
	db := setupTestDB(t)
	svcCtx := newTestSvcCtx(db)

	// 预插入用户
	db.Exec("INSERT INTO users (id, username, phone, password) VALUES (2, 'test2', '13800000002', '')")

	l := NewSmsLoginApiLogic(context.Background(), svcCtx)

	// 未注册手机号
	_, err := l.SmsLogin(&types.SmsLoginRequest{Phone: "13999999999", Code: "any"})
	if err == nil {
		t.Fatal("unregistered phone should fail")
	}

	// 错误验证码
	_, err = l.SmsLogin(&types.SmsLoginRequest{Phone: "13800000002", Code: "wrong"})
	if err == nil {
		t.Fatal("wrong code should fail")
	}
}
