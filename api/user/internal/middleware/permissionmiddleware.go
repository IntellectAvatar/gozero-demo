package middleware

import (
	"net/http"
	"strings"

	"gozero-demo/internal/response"

	"github.com/zeromicro/go-zero/core/logx"
)

// PermissionMiddleware 权限校验中间件 — 检查 JWT claims 中的 permissions
// requiredPerm: 需要的权限 code，如 "user:write"
func PermissionMiddleware(requiredPerm string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			perms, ok := r.Context().Value("permissions").([]string)
			if !ok {
				// go-zero 的 JWT Authorize 存储 claims，但切片类型需要特殊处理
				perms = getPermissionsFromContext(r)
			}

			for _, p := range perms {
				if p == requiredPerm {
					next(w, r)
					return
				}
			}

			logx.WithContext(r.Context()).Errorf("权限不足: 需要 %s, 拥有 %v", requiredPerm, perms)
			response.Fail(w, r, http.StatusForbidden, response.CodeForbidden, "forbidden")
		}
	}
}

// getPermissionsFromContext 从 go-zero JWT claims context 中提取 permissions 切片
func getPermissionsFromContext(r *http.Request) []string {
	// go-zero handler.Authorize stores claims as context values
	// 尝试多种类型
	v := r.Context().Value("permissions")
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		perms := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				perms = append(perms, s)
			}
		}
		return perms
	}
	return nil
}

// HasPermission 检查 context 中是否包含指定权限（供 logic 层使用）
func HasPermission(perms []string, required string) bool {
	for _, p := range perms {
		if strings.EqualFold(p, required) {
			return true
		}
	}
	return false
}
