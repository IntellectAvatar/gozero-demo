package handler

import (
	"net/http"

	"gozero-demo/api/user/internal/logic"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/internal/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		if err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, resp, "register_success")
	}
}
