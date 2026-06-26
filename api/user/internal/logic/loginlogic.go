package logic

import (
	"context"
	"time"

	"gozero-demo/api/user/internal/svc"
	"gozero-demo/api/user/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	// 从数据库查询用户
	type User struct {
		ID       int64  `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Password string `gorm:"column:password"`
	}
	var user User
	if err := l.svcCtx.DB.Table("users").Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errInvalidCredential
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errInvalidCredential
	}

	l.Infof("用户登录成功: %s (id=%d)", user.Username, user.ID)

	// 查询用户的角色和权限
	roles, perms := l.getUserRolesPermissions(user.ID)

	// 生成 JWT（包含角色和权限）
	now := time.Now().Unix()
	expire := now + l.svcCtx.Config.Auth.AccessExpire
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":      user.ID,
		"username":    user.Username,
		"roles":       roles,
		"permissions": perms,
		"iat":         now,
		"exp":         expire,
	})
	tokenString, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{Token: tokenString, Expire: expire}, nil
}

var errInvalidCredential = errInvalidCredentialError{}

// getUserRolesPermissions 查询用户角色和权限（直查 DB）
func (l *LoginLogic) getUserRolesPermissions(userId int64) ([]string, []string) {
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

type errInvalidCredentialError struct{}

func (e errInvalidCredentialError) Error() string { return "用户名或密码错误" }
