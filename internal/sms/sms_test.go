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
