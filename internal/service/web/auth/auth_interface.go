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
	Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error)
	ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) (*response.MessageResponse, error)
	SendCode(ctx context.Context, req *request.SendCodeRequest) (*response.MessageResponse, error)
	GetBaseInfo(ctx context.Context, userID uint) (*memberResp.MemberBaseInfo, error)
	Logout(ctx context.Context, userID uint) error
}
