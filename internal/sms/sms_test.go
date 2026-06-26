package sms

import (
	"testing"
	"time"
)

func TestSendAndVerify(t *testing.T) {
	s := &CodeStore{codes: make(map[string]codeEntry)}

	if err := s.Send("13800138000"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	code := s.GetCode("13800138000")
	if len(code) != 6 {
		t.Errorf("code length = %d, want 6", len(code))
	}

	if !s.Verify("13800138000", code) {
		t.Error("Verify should succeed with correct code")
	}
	// 验证码消费：同一 code 不能重复使用
	if s.Verify("13800138000", code) {
		t.Error("Verify should fail after code is consumed")
	}
	if s.GetCode("13800138000") != "" {
		t.Error("GetCode should return empty after verification consumption")
	}
	if s.Verify("13800138000", "000000") {
		t.Error("Verify should fail with wrong code")
	}
}

func TestSendRateLimit(t *testing.T) {
	s := &CodeStore{codes: make(map[string]codeEntry)}
	s.Send("13800138000")
	err := s.Send("13800138000")
	if err == nil {
		t.Error("second send within 60s should fail")
	}
}

func TestCodeExpiry(t *testing.T) {
	s := &CodeStore{codes: make(map[string]codeEntry)}
	s.Send("13800138000")
	code := s.GetCode("13800138000")
	// 手动过期
	s.mu.Lock()
	entry := s.codes["13800138000"]
	entry.expireAt = time.Now().Add(-1 * time.Second)
	s.codes["13800138000"] = entry
	s.mu.Unlock()

	if s.Verify("13800138000", code) {
		t.Error("expired code should fail verification")
	}
}

func TestPhoneValidation(t *testing.T) {
	s := &CodeStore{codes: make(map[string]codeEntry)}

	tests := []struct {
		phone string
		want  bool // true = should succeed
	}{
		{"13800138000", true},
		{"23800138000", false},  // 不以 1 开头
		{"1380013800", false},   // 10 位
		{"138001380000", false}, // 12 位
		{"", false},             // 空
		{"13800abc000", false},  // 含字母
		{"03800138000", false},  // 以 0 开头
	}

	for _, tc := range tests {
		err := s.Send(tc.phone)
		if tc.want && err != nil {
			t.Errorf("Send(%q) should succeed, got: %v", tc.phone, err)
		}
		if !tc.want && err == nil {
			t.Errorf("Send(%q) should fail", tc.phone)
		}
	}
}

func TestExpiredEntryCleanup(t *testing.T) {
	s := &CodeStore{codes: make(map[string]codeEntry)}

	s.Send("13800138000")
	s.Send("13900139000")

	// 手动过期第一条
	s.mu.Lock()
	entry := s.codes["13800138000"]
	entry.expireAt = time.Now().Add(-1 * time.Second)
	s.codes["13800138000"] = entry
	s.mu.Unlock()

	// 发送新号码触发过期清理
	s.Send("13700137000")

	// 过期条目应被清理，有效条目应保留
	s.mu.RLock()
	_, expiredExists := s.codes["13800138000"]
	_, validExists := s.codes["13900139000"]
	s.mu.RUnlock()

	if expiredExists {
		t.Error("expired entry should have been cleaned up")
	}
	if !validExists {
		t.Error("valid entry should not be cleaned up")
	}
}
