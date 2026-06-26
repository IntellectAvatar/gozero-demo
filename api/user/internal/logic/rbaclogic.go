package logic

import (
	"context"

	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

// ====================== AssignRole ======================

type AssignRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleLogic {
	return &AssignRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AssignRoleLogic) AssignRole(req *types.AssignRoleRequest) error {
	if l.svcCtx.UserRpc != nil {
		_, err := l.svcCtx.UserRpc.AssignRole(l.ctx, &pb.AssignRoleRequest{
			UserId: req.UserID,
			RoleId: req.RoleID,
		})
		return err
	}
	l.svcCtx.DB.Exec("INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", req.UserID, req.RoleID)
	return nil
}

// ====================== RemoveRole ======================

type RemoveRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveRoleLogic {
	return &RemoveRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RemoveRoleLogic) RemoveRole(req *types.RemoveRoleRequest) error {
	if l.svcCtx.UserRpc != nil {
		_, err := l.svcCtx.UserRpc.RemoveRole(l.ctx, &pb.RemoveRoleRequest{
			UserId: req.UserID,
			RoleId: req.RoleID,
		})
		return err
	}
	l.svcCtx.DB.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", req.UserID, req.RoleID)
	return nil
}

// ====================== ListRoles ======================

type ListRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListRolesLogic) ListRoles() ([]types.RoleInfo, error) {
	if l.svcCtx.UserRpc != nil {
		resp, err := l.svcCtx.UserRpc.ListRoles(l.ctx, &pb.ListRolesRequest{})
		if err != nil {
			return nil, err
		}
		roles := make([]types.RoleInfo, 0, len(resp.Roles))
		for _, r := range resp.Roles {
			roles = append(roles, types.RoleInfo{ID: r.Id, Name: r.Name, Code: r.Code, Description: r.Description})
		}
		return roles, nil
	}
	// 降级直查
	type Role struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Code string `gorm:"column:code"`
		Desc string `gorm:"column:description"`
	}
	var rows []Role
	l.svcCtx.DB.Table("roles").Find(&rows)
	roles := make([]types.RoleInfo, 0, len(rows))
	for _, r := range rows {
		roles = append(roles, types.RoleInfo{ID: r.ID, Name: r.Name, Code: r.Code, Description: r.Desc})
	}
	return roles, nil
}

// ====================== ListPermissions ======================

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListPermissionsLogic) ListPermissions() ([]types.PermissionInfo, error) {
	if l.svcCtx.UserRpc != nil {
		resp, err := l.svcCtx.UserRpc.ListPermissions(l.ctx, &pb.ListPermissionsRequest{})
		if err != nil {
			return nil, err
		}
		perms := make([]types.PermissionInfo, 0, len(resp.Permissions))
		for _, p := range resp.Permissions {
			perms = append(perms, types.PermissionInfo{ID: p.Id, Code: p.Code, Name: p.Name, Description: p.Description})
		}
		return perms, nil
	}
	type Perm struct {
		ID   int64  `gorm:"column:id"`
		Code string `gorm:"column:code"`
		Name string `gorm:"column:name"`
		Desc string `gorm:"column:description"`
	}
	var rows []Perm
	l.svcCtx.DB.Table("permissions").Find(&rows)
	perms := make([]types.PermissionInfo, 0, len(rows))
	for _, p := range rows {
		perms = append(perms, types.PermissionInfo{ID: p.ID, Code: p.Code, Name: p.Name, Description: p.Desc})
	}
	return perms, nil
}

// ====================== AssignRolePermission ======================

type AssignRolePermissionLogic struct {
	logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext
}
func NewAssignRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermissionLogic {
	return &AssignRolePermissionLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *AssignRolePermissionLogic) AssignRolePermission(req *types.AssignRolePermissionRequest) error {
	if l.svcCtx.UserRpc != nil {
		_, err := l.svcCtx.UserRpc.AssignRolePermission(l.ctx, &pb.AssignRolePermissionRequest{RoleId: req.RoleID, PermissionId: req.PermissionID})
		return err
	}
	l.svcCtx.DB.Exec("INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", req.RoleID, req.PermissionID)
	return nil
}

type RemoveRolePermissionLogic struct {
	logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext
}
func NewRemoveRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveRolePermissionLogic {
	return &RemoveRolePermissionLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}
func (l *RemoveRolePermissionLogic) RemoveRolePermission(req *types.AssignRolePermissionRequest) error {
	if l.svcCtx.UserRpc != nil {
		_, err := l.svcCtx.UserRpc.RemoveRolePermission(l.ctx, &pb.RemoveRolePermissionRequest{RoleId: req.RoleID, PermissionId: req.PermissionID})
		return err
	}
	l.svcCtx.DB.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", req.RoleID, req.PermissionID)
	return nil
}
