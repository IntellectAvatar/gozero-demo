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

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// 检查用户名是否已存在
	var count int64
	l.svcCtx.DB.Model(&model.User{}).Where("username = ?", in.Username).Count(&count)
	if count > 0 {
		return nil, status.Error(codes.AlreadyExists, "用户名已存在")
	}

	// bcrypt 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "密码加密失败")
	}

	user := model.User{
		Username: in.Username,
		Password: string(hashed),
		Email:    in.Email,
	}

	if err := l.svcCtx.DB.Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, "创建用户失败")
	}

	l.Infof("新用户注册: %s (id=%d)", user.Username, user.ID)

	return &pb.CreateUserResponse{
		User: &pb.UserInfo{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		},
	}, nil
}
