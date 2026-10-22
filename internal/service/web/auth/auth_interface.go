package auth

import (
	"context"
	request "tuyou/internal/model/dto/request/web/auth"
	response "tuyou/internal/model/dto/response/web/auth"
	memberResp "tuyou/internal/model/dto/response/web/member"
)

// WebAuthService 认证服务接口
type WebAuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*response.MessageResponse, error)
	GetCaptcha(ctx context.Context) (*response.CaptchaResp, error)
	Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*response.RefreshResponse, error)
	ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) (*response.MessageResponse, error)
	SendCode(ctx context.Context, req *request.SendCodeRequest) (*response.MessageResponse, error)
	GetBaseInfo(ctx context.Context, userID uint) (*memberResp.MemberBaseInfo, error)
	Logout(ctx context.Context, userID uint, jti string) error
	ChangePhone(ctx context.Context, userID uint, req *request.ChangePhoneRequest) (*response.MessageResponse, error)
	UpdatePassword(ctx context.Context, userID uint, req *request.UpdatePasswordRequest) (*response.MessageResponse, error)

	// Cancel 注销账号（密码 + 短信验证码双重校验，成功后 status=2 并删除登录态）
	Cancel(ctx context.Context, userID uint, req *request.CancelRequest) (*response.MessageResponse, error)
}
