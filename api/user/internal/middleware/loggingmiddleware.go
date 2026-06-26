package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/timex"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// statusResponseWriter 包装 http.ResponseWriter，捕获状态码。
// 首次调用 WriteHeader 时会记录状态码（默认 200）。
type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		code:           http.StatusOK,
	}
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

// Flush 实现 http.Flusher 接口，支持 SSE 等流式响应。
func (w *statusResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现 http.Hijacker 接口，支持 WebSocket 等协议升级。
func (w *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacked.Hijack()
	}
	return nil, nil, errors.New("server doesn't support hijacking")
}

// Unwrap 支持 http.ResponseController 解包。
func (w *statusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// LoggingMiddleware 是一个记录每次请求详细信息的中间件。
// 记录内容包括：请求方法、路径、状态码、耗时、客户端地址、User-Agent、请求体。
// 根据状态码自动选择日志级别：5xx → Error，4xx → Info，2xx/3xx → Info。
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := timex.Now()

		// 包装 ResponseWriter 以捕获状态码
		sw := newStatusResponseWriter(w)

		// 读取请求体（最多 1024 字节），同时保留原始 Body 供下游 handler 使用
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(io.LimitReader(r.Body, 1024))
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		next(sw, r)

		duration := timex.Since(start)
		logger := logx.WithContext(r.Context()).WithDuration(duration)

		code := sw.code
		fields := []logx.LogField{
			logx.Field("method", r.Method),
			logx.Field("path", r.RequestURI),
			logx.Field("statusCode", code),
			logx.Field("duration", duration.String()),
			logx.Field("remoteAddr", httpx.GetRemoteAddr(r)),
		}
		if ua := r.UserAgent(); ua != "" {
			fields = append(fields, logx.Field("userAgent", ua))
		}
		if len(reqBody) > 0 {
			fields = append(fields, logx.Field("requestBody", string(reqBody)))
		}

		switch {
		case code >= http.StatusInternalServerError:
			logger.WithFields(fields...).Error("请求处理完成")
		default:
			logger.WithFields(fields...).Info("请求处理完成")
		}
	}
}
