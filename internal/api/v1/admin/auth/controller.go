package auth

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/admin/auth"
	"tuyou/internal/service/admin/auth"
	"tuyou/pkg/response"
)

// AdminHandler 认证处理器
type AdminHandler struct {
	authSvc auth.AdminAuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc auth.AdminAuthService) *AdminHandler {
	return &AdminHandler{authSvc: authSvc}
}

// GetCaptcha 获取图形验证码
func (h *AdminHandler) GetCaptcha(c *gin.Context) {
	resp, err := h.authSvc.GetCaptcha(c.Request.Context())
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}

// Login 登录
func (h *AdminHandler) Login(c *gin.Context) {
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

// Me 获取当前登录管理员信息
func (h *AdminHandler) Me(c *gin.Context) {
	resp, err := h.authSvc.GetProfile(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}

// UpdatePassword 修改当前登录管理员密码
func (h *AdminHandler) UpdatePassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数校验失败")
		return
	}
	if err := h.authSvc.UpdatePassword(c.Request.Context(), middleware.GetUserID(c), &req); err != nil {
		response.BizError(c, err)
		return
	}
	response.SuccessWithMessage(c, "密码修改成功", nil)
}
