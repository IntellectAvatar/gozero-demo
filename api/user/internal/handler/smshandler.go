package handler

import (
	"net/http"

	"gozero-demo/api/user/internal/logic"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/internal/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SendSmsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSmsRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		if err := logic.NewSendSmsApiLogic(r.Context(), svcCtx).SendSms(&req); err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, nil, "success")
	}
}

func SmsRegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsRegisterRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewSmsRegisterApiLogic(r.Context(), svcCtx)
		resp, err := l.SmsRegister(&req)
		if err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, resp, "register_success")
	}
}

func SmsLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsLoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewSmsLoginApiLogic(r.Context(), svcCtx)
		resp, err := l.SmsLogin(&req)
		if err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, resp, "login_success")
	}
}
