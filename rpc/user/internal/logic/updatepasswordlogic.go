package logic

import (
	"context"

	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdatePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(in *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	var user model.User
	if err := l.svcCtx.DB.First(&user, in.Id).Error; err != nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	// 校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.OldPassword)); err != nil {
		return nil, status.Error(codes.PermissionDenied, "旧密码错误")
	}

	// 加密新密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "密码加密失败")
	}

	if err := l.svcCtx.DB.Model(&user).Update("password", string(hashed)).Error; err != nil {
		return nil, status.Error(codes.Internal, "密码更新失败")
	}

	l.Infof("用户密码已更新: id=%d", user.ID)
	return &pb.UpdatePasswordResponse{}, nil
}
