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

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserInfo 查询当前登录用户信息，userId 从 JWT token 中获取
func (l *UserInfoLogic) UserInfo() (*types.UserInfoResponse, error) {
	// 从 JWT claims 中取 userId，不信任请求参数
	userId := middleware.GetJwtClaimInt64(l.ctx, "userId")
	if userId == 0 {
		return nil, status.Error(codes.Unauthenticated, "无法获取用户信息")
	}

	l.Infof("查询用户信息: userId=%d", userId)

	// 优先调 user-rpc
	if l.svcCtx.UserRpc != nil {
		userResp, rpcErr := l.svcCtx.UserRpc.GetUser(l.ctx, &pb.GetUserRequest{Id: userId})
		if rpcErr != nil {
			l.Errorf("调用 user rpc 失败: %v", rpcErr)
			return l.fallback(userId)
		}
		return &types.UserInfoResponse{
			ID:       userResp.User.Id,
			Username: userResp.User.Username,
			Email:    userResp.User.Email,
		}, nil
	}

	return l.fallback(userId)
}

func (l *UserInfoLogic) fallback(userId int64) (*types.UserInfoResponse, error) {
	l.Infof("使用降级模式查询用户: userId=%d", userId)

	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Email    string `gorm:"column:email"`
	}
	var user User
	if err := l.svcCtx.DB.Table("users").Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}
	return &types.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
