package handler

import (
	"net/http"
	"strings"

	"gozero-demo/internal/response"
)

// handleError 根据错误信息映射到统一错误响应
func handleError(w http.ResponseWriter, r *http.Request, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "用户名或密码错误"):
		response.Fail(w, r, http.StatusUnauthorized, response.CodeWrongPassword, "wrong_password")
	case strings.Contains(msg, "旧密码错误"):
		response.Fail(w, r, http.StatusForbidden, response.CodeOldPasswordWrong, "old_password_wrong")
	case strings.Contains(msg, "用户名已存在"):
		response.Fail(w, r, http.StatusConflict, response.CodeUserExists, "user_exists")
	case strings.Contains(msg, "用户不存在"):
		response.Fail(w, r, http.StatusNotFound, response.CodeUserNotFound, "user_not_found")
	case strings.Contains(msg, "无法获取用户信息") || strings.Contains(msg, "token"):
		response.Fail(w, r, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
	case strings.Contains(msg, "不可用") || strings.Contains(msg, "Unavailable"):
		response.Fail(w, r, http.StatusServiceUnavailable, response.CodeServiceUnavailable, "service_unavailable")
	default:
		response.Fail(w, r, http.StatusInternalServerError, response.CodeInternalError, "internal_error")
	}
}
