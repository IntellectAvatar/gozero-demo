package logic

import (
	"context"

	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUsersLogic) ListUsers(req *types.ListUsersRequest) ([]any, int64, error) {
	if l.svcCtx.UserRpc != nil {
		resp, rpcErr := l.svcCtx.UserRpc.ListUsers(l.ctx, &pb.ListUsersRequest{
			Page:     req.Page,
			PageSize: req.PageSize,
		})
		if rpcErr != nil {
			l.Errorf("rpc ListUsers 失败，降级直查 DB: %v", rpcErr)
			return l.fallbackList(req)
		}
		users := make([]any, 0, len(resp.Users))
		for _, u := range resp.Users {
			users = append(users, types.UserInfoResponse{
				ID:       u.Id,
				Username: u.Username,
				Email:    u.Email,
			})
		}
		return users, resp.Total, nil
	}
	return l.fallbackList(req)
}

func (l *ListUsersLogic) fallbackList(req *types.ListUsersRequest) ([]any, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Email    string `gorm:"column:email"`
	}

	var total int64
	l.svcCtx.DB.Table("users").Count(&total)

	var rows []User
	offset := (page - 1) * pageSize
	l.svcCtx.DB.Table("users").Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&rows)

	users := make([]any, 0, len(rows))
	for _, u := range rows {
		users = append(users, types.UserInfoResponse{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
		})
	}
	l.Infof("用户列表查询（降级模式）: page=%d, total=%d", page, total)
	return users, total, nil
}
