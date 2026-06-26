package sms

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type CodeStore struct {
	mu    sync.RWMutex
	codes map[string]codeEntry // phone -> {code, expire, lastSent}
}

type codeEntry struct {
	code     string
	expireAt time.Time
	lastSent time.Time
}

var DefaultStore = &CodeStore{codes: make(map[string]codeEntry)}

const sendInterval = 60 * time.Second  // 同一手机号发送间隔
const codeExpire = 5 * time.Minute     // 验证码有效期
const codeLen = 6

// Send 生成验证码并存储，返回 code（生产环境不发 code）
func (s *CodeStore) Send(phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.codes[phone]; ok {
		if time.Since(entry.lastSent) < sendInterval {
			return fmt.Errorf("发送过于频繁，请 %d 秒后再试", int(sendInterval.Seconds()))
		}
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	s.codes[phone] = codeEntry{
		code:     code,
		expireAt: time.Now().Add(codeExpire),
		lastSent: time.Now(),
	}
	return nil
}

// Verify 校验验证码
func (s *CodeStore) Verify(phone, code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.codes[phone]
	if !ok {
		return false
	}
	if time.Now().After(entry.expireAt) {
		return false
	}
	return entry.code == code
}

// GetCode 获取验证码（仅测试用）
func (s *CodeStore) GetCode(phone string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if entry, ok := s.codes[phone]; ok {
		if time.Now().Before(entry.expireAt) {
			return entry.code
		}
	}
	return ""
}
