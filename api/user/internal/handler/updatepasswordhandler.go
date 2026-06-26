package handler

import (
	"net/http"

	"gozero-demo/api/user/internal/logic"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/internal/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdatePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdatePasswordRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewUpdatePasswordLogic(r.Context(), svcCtx)
		if err := l.UpdatePassword(&req); err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, nil, "password_changed")
	}
}
