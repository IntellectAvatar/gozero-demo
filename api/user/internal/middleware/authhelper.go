package middleware

import (
	"context"
	"encoding/json"
)

// GetJwtClaim 从 context 中获取 JWT 自定义声明值。
func GetJwtClaim(ctx context.Context, key string) any {
	return ctx.Value(key)
}

// GetJwtClaimString 从 context 中获取 JWT 自定义声明值并转为字符串。
func GetJwtClaimString(ctx context.Context, key string) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetJwtClaimInt64 从 context 中获取 JWT 自定义声明值并转为 int64。
// JWT 中的数字可能是 json.Number 或 float64 类型。
func GetJwtClaimInt64(ctx context.Context, key string) int64 {
	v := ctx.Value(key)
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case json.Number:
		i, _ := n.Int64()
		return i
	case float64:
		return int64(n)
	case int64:
		return n
	}
	return 0
}
