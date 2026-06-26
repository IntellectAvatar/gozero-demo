// Package i18n 提供中英文国际化消息
package i18n

import "net/http"

// Lang 语言类型
type Lang string

const (
	ZH Lang = "zh-CN"
	EN Lang = "en-US"
)

// Detect 从 Accept-Language header 检测语言
func Detect(r *http.Request) Lang {
	h := r.Header.Get("Accept-Language")
	if len(h) >= 2 && h[:2] == "zh" {
		return ZH
	}
	return EN
}

// Msg 获取多语言消息，key 不存在时返回 key 本身
func Msg(lang Lang, key string) string {
	if m, ok := Messages[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := enMessages[key]; ok {
		return v
	}
	return key
}

// Messages 多语言消息表
var Messages = map[Lang]map[string]string{
	ZH: zhMessages,
	EN: enMessages,
}

var zhMessages = map[string]string{
	"success":             "成功",
	"error":               "错误",
	"invalid_request":     "请求参数无效",
	"unauthorized":        "未授权，请先登录",
	"not_found":           "资源不存在",
	"too_many_requests":   "请求过于频繁，请稍后再试",
	"internal_error":      "服务器内部错误",
	"service_unavailable": "服务暂不可用，已启用降级模式",
	"user_not_found":      "用户不存在",
	"user_exists":         "用户名已存在",
	"wrong_password":      "用户名或密码错误",
	"old_password_wrong":  "旧密码错误",
	"password_changed":    "密码修改成功",
	"register_success":    "注册成功",
	"login_success":       "登录成功",
}

var enMessages = map[string]string{
	"success":             "Success",
	"error":               "Error",
	"invalid_request":     "Invalid request parameters",
	"unauthorized":        "Unauthorized, please login first",
	"not_found":           "Resource not found",
	"too_many_requests":   "Too many requests, please try again later",
	"internal_error":      "Internal server error",
	"service_unavailable": "Service temporarily unavailable, fallback enabled",
	"user_not_found":      "User not found",
	"user_exists":         "Username already exists",
	"wrong_password":      "Invalid username or password",
	"old_password_wrong":  "Old password is incorrect",
	"password_changed":    "Password changed successfully",
	"register_success":    "Registration successful",
	"login_success":       "Login successful",
}
