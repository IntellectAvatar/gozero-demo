package main

import (
	"context"
	"flag"
	"fmt"

	"gozero-demo/rpc/user/internal/config"
	"gozero-demo/rpc/user/internal/logic"
	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var (
	configFile = flag.String("f", "etc/user-rpc.yaml", "the config file")
	version    = "dev"
)

type UserServer struct {
	pb.UnimplementedUserServer
	svcCtx *svc.ServiceContext
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return logic.NewGetUserLogic(ctx, s.svcCtx).GetUser(req)
}
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return logic.NewListUsersLogic(ctx, s.svcCtx).ListUsers(req)
}
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return logic.NewCreateUserLogic(ctx, s.svcCtx).CreateUser(req)
}
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return logic.NewUpdateUserLogic(ctx, s.svcCtx).UpdateUser(req)
}
func (s *UserServer) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	return logic.NewUpdatePasswordLogic(ctx, s.svcCtx).UpdatePassword(req)
}

// RBAC
func (s *UserServer) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	return logic.NewListRolesLogic(ctx, s.svcCtx).ListRoles(req)
}
func (s *UserServer) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	return logic.NewAssignRoleLogic(ctx, s.svcCtx).AssignRole(req)
}
func (s *UserServer) RemoveRole(ctx context.Context, req *pb.RemoveRoleRequest) (*pb.RemoveRoleResponse, error) {
	return logic.NewRemoveRoleLogic(ctx, s.svcCtx).RemoveRole(req)
}
func (s *UserServer) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	return logic.NewGetUserRolesLogic(ctx, s.svcCtx).GetUserRoles(req)
}
func (s *UserServer) ListPermissions(ctx context.Context, req *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	return logic.NewListPermissionsLogic(ctx, s.svcCtx).ListPermissions(req)
}
func (s *UserServer) AssignRolePermission(ctx context.Context, req *pb.AssignRolePermissionRequest) (*pb.AssignRolePermissionResponse, error) {
	return logic.NewAssignRolePermissionLogic(ctx, s.svcCtx).AssignRolePermission(req)
}
func (s *UserServer) RemoveRolePermission(ctx context.Context, req *pb.RemoveRolePermissionRequest) (*pb.RemoveRolePermissionResponse, error) {
	return logic.NewRemoveRolePermissionLogic(ctx, s.svcCtx).RemoveRolePermission(req)
}

// SMS
func (s *UserServer) SendSms(ctx context.Context, req *pb.SendSmsRequest) (*pb.SendSmsResponse, error) {
	return logic.NewSendSmsLogic(ctx, s.svcCtx).SendSms(req)
}
func (s *UserServer) SmsRegister(ctx context.Context, req *pb.SmsRegisterRequest) (*pb.SmsRegisterResponse, error) {
	return logic.NewSmsRegisterLogic(ctx, s.svcCtx).SmsRegister(req)
}
func (s *UserServer) SmsLogin(ctx context.Context, req *pb.SmsLoginRequest) (*pb.SmsLoginResponse, error) {
	return logic.NewSmsLoginLogic(ctx, s.svcCtx).SmsLogin(req)
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)
	seedRBAC(svcCtx)

	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterUserServer(grpcServer, &UserServer{svcCtx: svcCtx})
	})
	defer server.Stop()

	fmt.Printf("Starting user-rpc %s at %s...\n", version, c.ListenOn)
	server.Start()
}

// seedRBAC 初始化默认角色和权限种子数据
func seedRBAC(ctx *svc.ServiceContext) {
	db := ctx.DB

	// 默认权限
	perms := []model.Permission{
		{Code: "user:read", Name: "查看用户"},
		{Code: "user:write", Name: "修改用户"},
		{Code: "user:delete", Name: "删除用户"},
		{Code: "role:read", Name: "查看角色"},
		{Code: "role:assign", Name: "分配角色"},
	}
	for i := range perms {
		db.Where("code = ?", perms[i].Code).FirstOrCreate(&perms[i])
	}

	// 默认角色
	adminRole := model.Role{Code: "admin", Name: "管理员", Description: "系统管理员"}
	userRole := model.Role{Code: "user", Name: "普通用户", Description: "普通用户"}
	db.Where("code = ?", "admin").FirstOrCreate(&adminRole)
	db.Where("code = ?", "user").FirstOrCreate(&userRole)

	// 给 admin 分配所有权限
	for _, p := range perms {
		db.Where("role_id = ? AND permission_id = ?", adminRole.ID, p.ID).
			FirstOrCreate(&model.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID})
	}
	// 给 user 分配只读权限
	db.Where("role_id = ? AND permission_id = ?", userRole.ID, perms[0].ID).
		FirstOrCreate(&model.RolePermission{RoleID: userRole.ID, PermissionID: perms[0].ID})

	logx.Infof("RBAC 种子数据初始化完成: %d 权限, %d 角色", len(perms), 2)
}
