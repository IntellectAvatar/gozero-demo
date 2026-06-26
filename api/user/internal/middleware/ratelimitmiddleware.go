package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	xrate "golang.org/x/time/rate"
)

// visitorInfo 记录每个 IP 的限流器和最后活跃时间。
type visitorInfo struct {
	limiter  *xrate.Limiter
	lastSeen time.Time
}

// RateLimiter 基于每 IP 令牌桶的限流器。
// 使用 golang.org/x/time/rate 实现，内存存储，定期清理不活跃的条目。
type RateLimiter struct {
	visitors map[string]*visitorInfo
	mu       sync.RWMutex
	rate     xrate.Limit
	burst    int
}

// NewRateLimiter 创建一个每 IP 限流器。
// rps: 每秒允许的请求数，burst: 突发请求数。
func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorInfo),
		rate:     xrate.Limit(rps),
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

// cleanup 定期清理超过 5 分钟未活跃的访客记录，防止内存泄漏。
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, info := range rl.visitors {
			if time.Since(info.lastSeen) > 5*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getLimiter 获取或创建指定 IP 的限流器。
func (rl *RateLimiter) getLimiter(ip string) *xrate.Limiter {
	rl.mu.RLock()
	info, ok := rl.visitors[ip]
	rl.mu.RUnlock()

	if ok {
		info.lastSeen = time.Now()
		return info.limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查，避免重复创建
	if info, ok := rl.visitors[ip]; ok {
		info.lastSeen = time.Now()
		return info.limiter
	}

	limiter := xrate.NewLimiter(rl.rate, rl.burst)
	rl.visitors[ip] = &visitorInfo{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	return limiter
}

// Handle 返回一个 rest.Middleware，对超过频率限制的 IP 返回 429 Too Many Requests。
func (rl *RateLimiter) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := httpx.GetRemoteAddr(r)
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			logx.WithContext(r.Context()).Errorf(
				"频率限制触发 - IP: %s, 路径: %s, 方法: %s",
				ip, r.URL.Path, r.Method,
			)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"code":429,"message":"请求过于频繁，请稍后再试"}`))
			return
		}
		next(w, r)
	}
}
