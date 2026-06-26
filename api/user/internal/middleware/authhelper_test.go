package middleware

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetJwtClaim(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userId", "test-user")
	if got := GetJwtClaim(ctx, "userId"); got != "test-user" {
		t.Errorf("GetJwtClaim = %v, want test-user", got)
	}
	if got := GetJwtClaim(ctx, "missing"); got != nil {
		t.Errorf("GetJwtClaim(missing) = %v, want nil", got)
	}
}

func TestGetJwtClaimString(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
		want  string
	}{
		{"string value", "name", "张三", "张三"},
		{"missing key", "missing", nil, ""},
		{"int value returns empty", "num", 123, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.value != nil {
				ctx = context.WithValue(ctx, tt.key, tt.value)
			}
			if got := GetJwtClaimString(ctx, tt.key); got != tt.want {
				t.Errorf("GetJwtClaimString = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGetJwtClaimInt64(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int64
	}{
		{"json.Number", json.Number("12345"), 12345},
		{"float64", float64(42), 42},
		{"int64", int64(99), 99},
		{"missing key", nil, 0},
		{"string returns 0", "abc", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.value != nil {
				ctx = context.WithValue(ctx, "userId", tt.value)
			}
			if got := GetJwtClaimInt64(ctx, "userId"); got != tt.want {
				t.Errorf("GetJwtClaimInt64 = %d, want %d", got, tt.want)
			}
		})
	}
}
