package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/admin/auth"
	"tuyou/internal/repository"
	"tuyou/internal/service/admin/auth"
	"tuyou/pkg/jwt"
	"tuyou/pkg/response"
)

// AdminHandler 认证处理器
type AdminHandler struct {
	authSvc   auth.AdminAuthService
	redisRepo repository.RedisRepository
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc auth.AdminAuthService, redisRepo repository.RedisRepository) *AdminHandler {
	return &AdminHandler{authSvc: authSvc, redisRepo: redisRepo}
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

	resp, err := h.authSvc.Login(c.Request.Context(), &req, c.Request.UserAgent())
	if err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, resp)
}

// Logout 退出登录
// POST /api/v1/admin/logout
// 从白名单删除当前 token 使其立即失效；幂等，token 已失效时也返回成功
func (h *AdminHandler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if claims, err := jwt.ParseToken(token); err == nil {
		_ = h.redisRepo.Del(c.Request.Context(), jwt.AdminTokenKey(claims.GetUserID(), token))
	}
	response.SuccessWithMessage(c, "退出登录成功", struct{}{})
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
