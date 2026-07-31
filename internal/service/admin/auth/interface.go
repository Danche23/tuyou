package auth

import (
	"context"
	request "tuyou/internal/model/dto/request/admin/auth"
	response "tuyou/internal/model/dto/response/admin/auth"
)

type AdminAuthService interface {
	// Login 登录
	Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error)

	// GetCaptcha 获取图形验证码
	GetCaptcha(ctx context.Context) (*response.CaptchaResponse, error)
}
