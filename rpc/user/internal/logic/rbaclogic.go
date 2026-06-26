package logic

import (
	"context"

	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ====================== ListRoles ======================

type ListRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListRolesLogic) ListRoles(in *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	var roles []model.Role
	l.svcCtx.DB.Find(&roles)
	pbRoles := make([]*pb.RoleInfo, 0, len(roles))
	for _, r := range roles {
		pbRoles = append(pbRoles, &pb.RoleInfo{
			Id:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
		})
	}
	return &pb.ListRolesResponse{Roles: pbRoles}, nil
}

// ====================== AssignRole ======================

type AssignRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleLogic {
	return &AssignRoleLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AssignRoleLogic) AssignRole(in *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	ur := model.UserRole{UserID: in.UserId, RoleID: in.RoleId}
	if err := l.svcCtx.DB.Save(&ur).Error; err != nil {
		return nil, status.Error(codes.Internal, "分配角色失败")
	}
	l.Infof("分配角色: userId=%d roleId=%d", in.UserId, in.RoleId)
	return &pb.AssignRoleResponse{}, nil
}

// ====================== RemoveRole ======================

type RemoveRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveRoleLogic {
	return &RemoveRoleLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RemoveRoleLogic) RemoveRole(in *pb.RemoveRoleRequest) (*pb.RemoveRoleResponse, error) {
	l.svcCtx.DB.Where("user_id = ? AND role_id = ?", in.UserId, in.RoleId).Delete(&model.UserRole{})
	l.Infof("移除角色: userId=%d roleId=%d", in.UserId, in.RoleId)
	return &pb.RemoveRoleResponse{}, nil
}

// ====================== GetUserRoles ======================

type GetUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetUserRolesLogic) GetUserRoles(in *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	var roleIDs []int64
	l.svcCtx.DB.Model(&model.UserRole{}).Where("user_id = ?", in.UserId).Pluck("role_id", &roleIDs)

	var roles []model.Role
	if len(roleIDs) > 0 {
		l.svcCtx.DB.Where("id IN ?", roleIDs).Find(&roles)
	}

	var permIDs []int64
	l.svcCtx.DB.Model(&model.RolePermission{}).Where("role_id IN ?", roleIDs).Pluck("permission_id", &permIDs)

	var perms []string
	if len(permIDs) > 0 {
		l.svcCtx.DB.Model(&model.Permission{}).Where("id IN ?", permIDs).Pluck("code", &perms)
	}

	pbRoles := make([]*pb.RoleInfo, 0, len(roles))
	for _, r := range roles {
		pbRoles = append(pbRoles, &pb.RoleInfo{
			Id:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
		})
	}
	return &pb.GetUserRolesResponse{Roles: pbRoles, Permissions: perms}, nil
}

// ====================== ListPermissions ======================

type ListPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListPermissionsLogic) ListPermissions(in *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	var perms []model.Permission
	l.svcCtx.DB.Find(&perms)
	pbPerms := make([]*pb.PermissionInfo, 0, len(perms))
	for _, p := range perms {
		pbPerms = append(pbPerms, &pb.PermissionInfo{
			Id:          p.ID,
			Code:        p.Code,
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return &pb.ListPermissionsResponse{Permissions: pbPerms}, nil
}

// ====================== AssignRolePermission ======================

type AssignRolePermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermissionLogic {
	return &AssignRolePermissionLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AssignRolePermissionLogic) AssignRolePermission(in *pb.AssignRolePermissionRequest) (*pb.AssignRolePermissionResponse, error) {
	rp := model.RolePermission{RoleID: in.RoleId, PermissionID: in.PermissionId}
	if err := l.svcCtx.DB.Save(&rp).Error; err != nil {
		return nil, status.Error(codes.Internal, "分配权限失败")
	}
	l.Infof("分配权限: roleId=%d permId=%d", in.RoleId, in.PermissionId)
	return &pb.AssignRolePermissionResponse{}, nil
}

// ====================== RemoveRolePermission ======================

type RemoveRolePermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveRolePermissionLogic {
	return &RemoveRolePermissionLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RemoveRolePermissionLogic) RemoveRolePermission(in *pb.RemoveRolePermissionRequest) (*pb.RemoveRolePermissionResponse, error) {
	l.svcCtx.DB.Where("role_id = ? AND permission_id = ?", in.RoleId, in.PermissionId).Delete(&model.RolePermission{})
	l.Infof("移除权限: roleId=%d permId=%d", in.RoleId, in.PermissionId)
	return &pb.RemoveRolePermissionResponse{}, nil
}
