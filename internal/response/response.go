package response

import (
	"encoding/json"
	"net/http"

	"gozero-demo/internal/i18n"
)

// Code 业务状态码
type Code int

const (
	CodeSuccess           Code = 0
	CodeBadRequest        Code = 40000
	CodeUnauthorized      Code = 40100
	CodeForbidden         Code = 40300
	CodeNotFound          Code = 40400
	CodeTooManyRequests   Code = 42900
	CodeInternalError     Code = 50000
	CodeServiceUnavailable Code = 50300

	CodeUserNotFound     Code = 10001
	CodeUserExists       Code = 10002
	CodeWrongPassword    Code = 10003
	CodeOldPasswordWrong Code = 10004
)

// Body 统一响应体 — Data 始终为数组，Total 用于分页
type Body struct {
	Code    Code   `json:"Code"`
	Message string `json:"Message"`
	Data    []any  `json:"Data"`
	Total   *int64 `json:"Total,omitempty"` // 仅分页接口使用
}

// Ok 成功响应 — 单个对象自动包装为数组
func Ok(w http.ResponseWriter, r *http.Request, data any) {
	lang := i18n.Detect(r)
	writeJSON(w, http.StatusOK, Body{
		Code:    CodeSuccess,
		Message: i18n.Msg(lang, "success"),
		Data:    wrapSlice(data),
	})
}

// OkList 分页响应 — Data 为数组，Total 为总数
func OkList(w http.ResponseWriter, r *http.Request, data []any, total int64) {
	lang := i18n.Detect(r)
	writeJSON(w, http.StatusOK, Body{
		Code:    CodeSuccess,
		Message: i18n.Msg(lang, "success"),
		Data:    data,
		Total:   &total,
	})
}

// OkMsg 成功响应（自定义消息 key）
func OkMsg(w http.ResponseWriter, r *http.Request, data any, msgKey string) {
	lang := i18n.Detect(r)
	writeJSON(w, http.StatusOK, Body{
		Code:    CodeSuccess,
		Message: i18n.Msg(lang, msgKey),
		Data:    wrapSlice(data),
	})
}

// Fail 错误响应
func Fail(w http.ResponseWriter, r *http.Request, httpStatus int, code Code, msgKey string) {
	lang := i18n.Detect(r)
	writeJSON(w, httpStatus, Body{
		Code:    code,
		Message: i18n.Msg(lang, msgKey),
		Data:    []any{},
	})
}

func writeJSON(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

// wrapSlice 把单个对象包装为数组，nil 返回空数组
func wrapSlice(data any) []any {
	if data == nil {
		return []any{}
	}
	return []any{data}
}
