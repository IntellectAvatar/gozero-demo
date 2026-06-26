package handler

import (
	"net/http"

	"gozero-demo/api/user/internal/middleware"
	"gozero-demo/api/user/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 公开路由
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/user/register", Handler: RegisterHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/user/login", Handler: LoginHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/user/sms/send", Handler: SendSmsHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/user/sms/register", Handler: SmsRegisterHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/user/sms/login", Handler: SmsLoginHandler(serverCtx)},
		},
	)

	// JWT 鉴权路由（普通用户可访问）
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/user/info", Handler: UserInfoHandler(serverCtx)},
			{Method: http.MethodPut, Path: "/user/update", Handler: UpdateUserHandler(serverCtx)},
			{Method: http.MethodPut, Path: "/user/password", Handler: UpdatePasswordHandler(serverCtx)},
			{Method: http.MethodGet, Path: "/user/list", Handler: ListUsersHandler(serverCtx)},
			{Method: http.MethodGet, Path: "/role/list", Handler: ListRolesHandler(serverCtx)},
			{Method: http.MethodGet, Path: "/permission/list", Handler: ListPermissionsHandler(serverCtx)},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	// 管理员路由（JWT + 权限校验: role:assign）
	server.AddRoutes(
		rest.WithMiddleware(
			middleware.PermissionMiddleware("role:assign"),
			rest.Route{Method: http.MethodPost, Path: "/role/assign", Handler: AssignRoleHandler(serverCtx)},
			rest.Route{Method: http.MethodPost, Path: "/role/remove", Handler: RemoveRoleHandler(serverCtx)},
			rest.Route{Method: http.MethodPost, Path: "/role/permission/assign", Handler: AssignRolePermissionHandler(serverCtx)},
			rest.Route{Method: http.MethodPost, Path: "/role/permission/remove", Handler: RemoveRolePermissionHandler(serverCtx)},
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
}
