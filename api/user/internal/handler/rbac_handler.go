package handler

import (
	"net/http"

	"gozero-demo/api/user/internal/logic"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/internal/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AssignRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssignRoleRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewAssignRoleLogic(r.Context(), svcCtx)
		if err := l.AssignRole(&req); err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, nil, "success")
	}
}

func RemoveRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoveRoleRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewRemoveRoleLogic(r.Context(), svcCtx)
		if err := l.RemoveRole(&req); err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, nil, "success")
	}
}

func ListRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListRolesLogic(r.Context(), svcCtx)
		roles, err := l.ListRoles()
		if err != nil {
			handleError(w, r, err)
			return
		}
		response.Ok(w, r, roles)
	}
}

func ListPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListPermissionsLogic(r.Context(), svcCtx)
		perms, err := l.ListPermissions()
		if err != nil {
			handleError(w, r, err)
			return
		}
		response.Ok(w, r, perms)
	}
}

func AssignRolePermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssignRolePermissionRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewAssignRolePermissionLogic(r.Context(), svcCtx)
		if err := l.AssignRolePermission(&req); err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, nil, "success")
	}
}

func RemoveRolePermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssignRolePermissionRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Fail(w, r, http.StatusBadRequest, response.CodeBadRequest, "invalid_request")
			return
		}
		l := logic.NewRemoveRolePermissionLogic(r.Context(), svcCtx)
		if err := l.RemoveRolePermission(&req); err != nil {
			handleError(w, r, err)
			return
		}
		response.OkMsg(w, r, nil, "success")
	}
}
