package logic

import (
	"context"

	"gozero-demo/api/user/internal/middleware"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserRequest) (*types.UserInfoResponse, error) {
	userId := middleware.GetJwtClaimInt64(l.ctx, "userId")
	if userId == 0 {
		return nil, status.Error(codes.Unauthenticated, "无法获取用户信息")
	}

	if l.svcCtx.UserRpc != nil {
		resp, rpcErr := l.svcCtx.UserRpc.UpdateUser(l.ctx, &pb.UpdateUserRequest{
			Id:    userId,
			Email: req.Email,
		})
		if rpcErr != nil {
			l.Errorf("rpc UpdateUser 失败，降级直写 DB: %v", rpcErr)
			return l.fallbackUpdate(userId, req.Email)
		}
		return &types.UserInfoResponse{
			ID:       resp.User.Id,
			Username: resp.User.Username,
			Email:    resp.User.Email,
		}, nil
	}
	return l.fallbackUpdate(userId, req.Email)
}

func (l *UpdateUserLogic) fallbackUpdate(userId int64, email string) (*types.UserInfoResponse, error) {
	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Email    string `gorm:"column:email"`
	}
	var user User
	if err := l.svcCtx.DB.Table("users").Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	if email != "" {
		l.svcCtx.DB.Table("users").Where("id = ?", userId).Update("email", email)
		user.Email = email
	}
	l.Infof("用户资料更新（降级模式）: userId=%d", userId)
	return &types.UserInfoResponse{ID: user.ID, Username: user.Username, Email: user.Email}, nil
}
