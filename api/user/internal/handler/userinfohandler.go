package handler

import (
	"net/http"

	"gozero-demo/api/user/internal/logic"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/internal/response"
)

func UserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.UserInfo()
		if err != nil {
			handleError(w, r, err)
			return
		}
		response.Ok(w, r, resp)
	}
}
