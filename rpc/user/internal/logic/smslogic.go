package logic

import (
	"context"

	"gozero-demo/internal/sms"
	"gozero-demo/rpc/user/internal/model"
	"gozero-demo/rpc/user/internal/svc"
	"gozero-demo/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ====================== SendSms ======================

type SendSmsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsLogic {
	return &SendSmsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendSmsLogic) SendSms(in *pb.SendSmsRequest) (*pb.SendSmsResponse, error) {
	if err := sms.DefaultStore.Send(in.Phone); err != nil {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}
	code := sms.DefaultStore.GetCode(in.Phone)
	l.Infof("发送验证码: phone=%s code=%s", in.Phone, code)
	return &pb.SendSmsResponse{Code: code}, nil
}

// ====================== SmsRegister ======================

type SmsRegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSmsRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsRegisterLogic {
	return &SmsRegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SmsRegisterLogic) SmsRegister(in *pb.SmsRegisterRequest) (*pb.SmsRegisterResponse, error) {
	if !sms.DefaultStore.Verify(in.Phone, in.Code) {
		return nil, status.Error(codes.InvalidArgument, "验证码错误或已过期")
	}
	// 检查手机号是否已注册
	var count int64
	l.svcCtx.DB.Model(&model.User{}).Where("phone = ?", in.Phone).Count(&count)
	if count > 0 {
		return nil, status.Error(codes.AlreadyExists, "手机号已注册")
	}
	// 创建用户（用户名自动生成）
	user := model.User{
		Username: "u_" + in.Phone,
		Password: "", // 验证码登录，暂无密码
		Phone:    in.Phone,
	}
	if err := l.svcCtx.DB.Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, "创建用户失败")
	}
	l.Infof("短信注册成功: phone=%s id=%d", in.Phone, user.ID)
	return &pb.SmsRegisterResponse{
		User: &pb.UserInfo{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// ====================== SmsLogin ======================

type SmsLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSmsLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SmsLoginLogic {
	return &SmsLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SmsLoginLogic) SmsLogin(in *pb.SmsLoginRequest) (*pb.SmsLoginResponse, error) {
	if !sms.DefaultStore.Verify(in.Phone, in.Code) {
		return nil, status.Error(codes.InvalidArgument, "验证码错误或已过期")
	}
	var user model.User
	if err := l.svcCtx.DB.Where("phone = ?", in.Phone).First(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "手机号未注册")
	}
	// 返回用户信息，由 api 层签发 JWT
	l.Infof("短信登录成功: phone=%s id=%d", in.Phone, user.ID)
	return &pb.SmsLoginResponse{}, nil
}
