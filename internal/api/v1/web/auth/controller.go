package auth

import (
	"github.com/gin-gonic/gin"
	request "tuyou/internal/model/dto/request/web/auth"
	"tuyou/internal/service/web/auth"
	"tuyou/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authSvc auth.WebAuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc auth.WebAuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	//校验密码
	if req.Password != req.ConfirmPassword {
		response.BadRequest(c, "两次输入的密码不一致")
		return
	}

	resp, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}

// SendCode 发送验证码
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req request.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.authSvc.SendCode(c.Request.Context(), &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}

// ResetPassword 重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.BadRequest(c, "两次输入密码不一致")
		return
	}
	resp, err := h.authSvc.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}
