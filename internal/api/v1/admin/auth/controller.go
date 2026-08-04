package auth

import (
	"github.com/gin-gonic/gin"
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
