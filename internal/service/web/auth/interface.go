package auth

import (
	"context"
	request "tuyou/internal/model/dto/request/web/auth"
	response "tuyou/internal/model/dto/response/web/auth"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*response.RegisterResponse, error)
	Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error)
	ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) (*response.MessageResponse, error)
}
