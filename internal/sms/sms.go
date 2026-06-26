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
	// 手机号格式校验: 11 位数字，1 开头
	if len(phone) != 11 || phone[0] != '1' {
		return fmt.Errorf("手机号格式错误")
	}
	for i := 0; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return fmt.Errorf("手机号格式错误")
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 清理所有过期条目
	now := time.Now()
	for p, entry := range s.codes {
		if now.After(entry.expireAt) {
			delete(s.codes, p)
		}
	}

	if entry, ok := s.codes[phone]; ok {
		if time.Since(entry.lastSent) < sendInterval {
			return fmt.Errorf("发送过于频繁，请 %d 秒后再试", int(sendInterval.Seconds()))
		}
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	s.codes[phone] = codeEntry{
		code:     code,
		expireAt: now.Add(codeExpire),
		lastSent: now,
	}
	return nil
}

// Verify 校验验证码，成功后删除条目防止重复使用
func (s *CodeStore) Verify(phone, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.codes[phone]
	if !ok {
		return false
	}
	if time.Now().After(entry.expireAt) {
		delete(s.codes, phone)
		return false
	}
	if entry.code == code {
		delete(s.codes, phone)
		return true
	}
	return false
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
