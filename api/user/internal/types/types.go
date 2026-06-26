package types

// ======================== 注册 ========================

type RegisterRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Email    string `json:"Email"`
}

type RegisterResponse struct {
	ID       int64  `json:"Id"`
	Username string `json:"Username"`
	Email    string `json:"Email"`
}

// ======================== 登录 ========================

type LoginRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type LoginResponse struct {
	Token  string `json:"Token"`
	Expire int64  `json:"Expire"`
}

// ======================== 用户信息 ========================

type UserInfoResponse struct {
	ID       int64  `json:"Id"`
	Username string `json:"Username"`
	Email    string `json:"Email"`
}

// ======================== 修改资料 ========================

type UpdateUserRequest struct {
	Email string `json:"Email"`
}

// ======================== 修改密码 ========================

type UpdatePasswordRequest struct {
	OldPassword string `json:"OldPassword"`
	NewPassword string `json:"NewPassword"`
}

// ======================== 用户列表 ========================

type ListUsersRequest struct {
	Page     int32 `json:"Page,default=1"`
	PageSize int32 `json:"PageSize,default=20"`
}

type ListUsersResponse struct {
	Users []UserInfoResponse `json:"Users"`
	Total int64              `json:"Total"`
}

// ======================== RBAC ========================

type AssignRoleRequest struct {
	UserID int64 `json:"UserId"`
	RoleID int64 `json:"RoleId"`
}

type RemoveRoleRequest struct {
	UserID int64 `json:"UserId"`
	RoleID int64 `json:"RoleId"`
}

type RoleInfo struct {
	ID          int64  `json:"Id"`
	Name        string `json:"Name"`
	Code        string `json:"Code"`
	Description string `json:"Description"`
}

type PermissionInfo struct {
	ID          int64  `json:"Id"`
	Code        string `json:"Code"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

type AssignRolePermissionRequest struct {
	RoleID       int64 `json:"RoleId"`
	PermissionID int64 `json:"PermissionId"`
}

// ======================== 手机验证码 ========================

type SendSmsRequest struct {
	Phone string `json:"Phone"`
}

type SmsLoginRequest struct {
	Phone string `json:"Phone"`
	Code  string `json:"Code"`
}

type SmsRegisterRequest struct {
	Phone string `json:"Phone"`
	Code  string `json:"Code"`
}
