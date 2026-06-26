package logic

import (
	"context"
	"errors"

	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	if l.svcCtx.UserRpc != nil {
		resp, rpcErr := l.svcCtx.UserRpc.CreateUser(l.ctx, &pb.CreateUserRequest{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		})
		if rpcErr != nil {
			l.Errorf("rpc CreateUser 失败，降级直写 DB: %v", rpcErr)
			return l.fallbackRegister(req)
		}
		return &types.RegisterResponse{
			ID:       resp.User.Id,
			Username: resp.User.Username,
			Email:    resp.User.Email,
		}, nil
	}
	return l.fallbackRegister(req)
}

func (l *RegisterLogic) fallbackRegister(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Password string `gorm:"column:password"`
		Email    string `gorm:"column:email"`
	}

	// 检查用户名是否已存在
	var count int64
	l.svcCtx.DB.Table("users").Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := User{
		Username: req.Username,
		Password: string(hashed),
		Email:    req.Email,
	}
	if err := l.svcCtx.DB.Table("users").Create(&user).Error; err != nil {
		return nil, err
	}

	// 如果是 GORM 返回的自增 ID
	if user.ID == 0 {
		l.svcCtx.DB.Table("users").Where("username = ?", req.Username).First(&user)
	}

	l.Infof("用户注册成功（降级模式）: %s (id=%d)", user.Username, user.ID)
	return &types.RegisterResponse{ID: user.ID, Username: user.Username, Email: user.Email}, nil
}
