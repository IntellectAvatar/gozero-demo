package logic

import (
	"context"

	"gozero-demo/api/user/internal/middleware"
	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdatePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(req *types.UpdatePasswordRequest) error {
	userId := middleware.GetJwtClaimInt64(l.ctx, "userId")
	if userId == 0 {
		return status.Error(codes.Unauthenticated, "无法获取用户信息")
	}

	if l.svcCtx.UserRpc != nil {
		_, rpcErr := l.svcCtx.UserRpc.UpdatePassword(l.ctx, &pb.UpdatePasswordRequest{
			Id:          userId,
			OldPassword: req.OldPassword,
			NewPassword: req.NewPassword,
		})
		if rpcErr != nil {
			l.Errorf("rpc UpdatePassword 失败，降级直写 DB: %v", rpcErr)
			return l.fallbackPassword(userId, req.OldPassword, req.NewPassword)
		}
		return nil
	}
	return l.fallbackPassword(userId, req.OldPassword, req.NewPassword)
}

func (l *UpdatePasswordLogic) fallbackPassword(userId int64, oldPwd, newPwd string) error {
	type User struct {
		Password string `gorm:"column:password"`
	}
	var user User
	if err := l.svcCtx.DB.Table("users").Where("id = ?", userId).First(&user).Error; err != nil {
		return status.Error(codes.NotFound, "用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPwd)); err != nil {
		return status.Error(codes.PermissionDenied, "旧密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	l.svcCtx.DB.Table("users").Where("id = ?", userId).Update("password", string(hashed))
	l.Infof("密码修改成功（降级模式）: userId=%d", userId)
	return nil
}
