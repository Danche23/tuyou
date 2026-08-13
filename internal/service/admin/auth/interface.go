package auth

import (
	"context"
	request "tuyou/internal/model/dto/request/admin/auth"
	response "tuyou/internal/model/dto/response/admin/auth"
)

type AdminAuthService interface {
	// Login 登录（userAgent 用于生成设备登录提醒文案）
	Login(ctx context.Context, req *request.LoginRequest, userAgent string) (*response.LoginResp, error)

	// GetCaptcha 获取图形验证码
	GetCaptcha(ctx context.Context) (*response.CaptchaResp, error)

	// GetProfile 获取当前登录管理员信息
	GetProfile(ctx context.Context, adminID uint) (*response.AdminInfo, error)

	// UpdatePassword 修改当前登录管理员密码
	UpdatePassword(ctx context.Context, adminID uint, req *request.ResetPasswordRequest) error
}
