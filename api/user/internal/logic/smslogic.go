package logic

import (
	"context"
	"time"

	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"
	"gozero-demo/internal/sms"
	"gozero-demo/rpc/user/pb"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

// ====================== SendSms ======================

type SendSmsApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendSmsApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsApiLogic {
	return &SendSmsApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendSmsApiLogic) SendSms(req *types.SendSmsRequest) error {
	// 优先调 rpc
	if l.svcCtx.UserRpc != nil {
		_, err := l.svcCtx.UserRpc.SendSms(l.ctx, &pb.SendSmsRequest{Phone: req.Phone})
		return err
	}
	return sms.DefaultStore.Send(req.Phone)
}

// ====================== SmsRegister ======================

type SmsRegisterApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSmsRegisterApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsRegisterApiLogic {
	return &SmsRegisterApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SmsRegisterApiLogic) SmsRegister(req *types.SmsRegisterRequest) (*types.RegisterResponse, error) {
	if l.svcCtx.UserRpc != nil {
		resp, err := l.svcCtx.UserRpc.SmsRegister(l.ctx, &pb.SmsRegisterRequest{Phone: req.Phone, Code: req.Code})
		if err != nil {
			return nil, err
		}
		return &types.RegisterResponse{
			ID:       resp.User.Id,
			Username: resp.User.Username,
			Email:    resp.User.Email,
		}, nil
	}
	// 降级：本地验证
	if !sms.DefaultStore.Verify(req.Phone, req.Code) {
		return nil, errInvalidCredential
	}
	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Email    string `gorm:"column:email"`
	}
	var count int64
	l.svcCtx.DB.Table("users").Where("phone = ?", req.Phone).Count(&count)
	if count > 0 {
		return nil, errInvalidCredential
	}
	user := User{Username: "u_" + req.Phone}
	l.svcCtx.DB.Table("users").Create(&user)
	return &types.RegisterResponse{ID: user.ID, Username: user.Username}, nil
}

// ====================== SmsLogin ======================

type SmsLoginApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSmsLoginApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsLoginApiLogic {
	return &SmsLoginApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SmsLoginApiLogic) SmsLogin(req *types.SmsLoginRequest) (*types.LoginResponse, error) {
	if l.svcCtx.UserRpc != nil {
		resp, err := l.svcCtx.UserRpc.SmsLogin(l.ctx, &pb.SmsLoginRequest{Phone: req.Phone, Code: req.Code})
		if err != nil {
			return nil, err
		}
		// rpc 验证成功，签发 JWT（包含角色和权限）
		return l.issueJWT(resp.UserId, resp.Username)
	}
	// 降级：本地验证 + DB + 签发 JWT
	if !sms.DefaultStore.Verify(req.Phone, req.Code) {
		return nil, errInvalidCredential
	}
	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
	}
	var user User
	if err := l.svcCtx.DB.Table("users").Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		return nil, errInvalidCredential
	}
	return l.issueJWT(user.ID, user.Username)
}

// issueJWT 签发 JWT，包含角色和权限
func (l *SmsLoginApiLogic) issueJWT(userId int64, username string) (*types.LoginResponse, error) {
	// 查询用户的角色和权限
	roles, perms := l.getUserRolesPermissions(userId)

	now := time.Now().Unix()
	expire := now + l.svcCtx.Config.Auth.AccessExpire
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":      userId,
		"username":    username,
		"roles":       roles,
		"permissions": perms,
		"iat":         now,
		"exp":         expire,
	})
	s, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return nil, err
	}
	return &types.LoginResponse{Token: s, Expire: expire}, nil
}

// getUserRolesPermissions 查询用户角色和权限（直查 DB）
func (l *SmsLoginApiLogic) getUserRolesPermissions(userId int64) ([]string, []string) {
	var roleIDs []int64
	l.svcCtx.DB.Table("user_roles").Where("user_id = ?", userId).Pluck("role_id", &roleIDs)
	if len(roleIDs) == 0 {
		return nil, nil
	}

	var roles []string
	l.svcCtx.DB.Table("roles").Where("id IN ?", roleIDs).Pluck("code", &roles)

	var permIDs []int64
	l.svcCtx.DB.Table("role_permissions").Where("role_id IN ?", roleIDs).Pluck("permission_id", &permIDs)

	var perms []string
	if len(permIDs) > 0 {
		l.svcCtx.DB.Table("permissions").Where("id IN ?", permIDs).Pluck("code", &perms)
	}
	return roles, perms
}

// 防止未使用 import
var _ = bcrypt.DefaultCost
