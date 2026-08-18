package auth

import (
	"gorm.io/gorm"
	"strings"

	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/admin/auth"
	"tuyou/internal/repository"
	"tuyou/internal/service/admin/auth"
	"tuyou/pkg/config"
	"tuyou/pkg/jwt"
	"tuyou/pkg/response"
)

// AdminController 认证处理器
type AdminController struct {
	authSvc   auth.AdminAuthService
	redisRepo repository.RedisRepository
	db        *gorm.DB
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc auth.AdminAuthService, redisRepo repository.RedisRepository, db *gorm.DB) *AdminController {
	return &AdminController{authSvc: authSvc, redisRepo: redisRepo, db: db}
}

// GetCaptcha 获取图形验证码
func (c *AdminController) GetCaptcha(context *gin.Context) {
	resp, err := c.authSvc.GetCaptcha(context.Request.Context())
	if err != nil {
		response.BizError(context, err)
		return
	}
	response.Success(context, resp)
}

// Login 登录
func (c *AdminController) Login(context *gin.Context) {
	var req request.LoginRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		response.BadRequest(context, err.Error())
		return
	}

	resp, err := c.authSvc.Login(context.Request.Context(), &req, context.Request.UserAgent())
	if err != nil {
		response.BizError(context, err)
		return
	}
	response.Success(context, resp)
}

// Logout 退出登录
// POST /api/v1/admin/logout
// 拉黑当前 jti 使其立即失效；幂等，token 已失效时也返回成功
func (c *AdminController) Logout(context *gin.Context) {
	token := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	if claims, err := jwt.ParseToken(token); err == nil {
		userID := claims.GetUserID()
		_ = c.redisRepo.Set(context.Request.Context(), jwt.BlacklistKey("admin", claims.ID), jwt.ReasonLogout, config.Get().JWT.ExpireHours)
		// 仅当会话标记仍是该 jti 时才删除，避免误删新登录的会话
		var cur string
		if err := c.redisRepo.Get(context.Request.Context(), jwt.AdminSessionKey(userID), &cur); err == nil && cur == claims.ID {
			_ = c.redisRepo.Del(context.Request.Context(), jwt.AdminSessionKey(userID))
		}
	}
	response.SuccessWithMessage(context, "退出登录成功", struct{}{})
}

// Me 获取当前登录管理员信息
func (c *AdminController) Me(context *gin.Context) {
	resp, err := c.authSvc.GetProfile(context.Request.Context(), middleware.GetUserID(context))
	if err != nil {
		response.BizError(context, err)
		return
	}
	response.Success(context, resp)
}

// UpdatePassword 修改当前登录管理员密码
func (c *AdminController) UpdatePassword(context *gin.Context) {
	var req request.ResetPasswordRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		response.BadRequest(context, "参数校验失败")
		return
	}
	if err := c.authSvc.UpdatePassword(context.Request.Context(), middleware.GetUserID(context), &req); err != nil {
		response.BizError(context, err)
		return
	}
	response.SuccessWithMessage(context, "密码修改成功", nil)
}
