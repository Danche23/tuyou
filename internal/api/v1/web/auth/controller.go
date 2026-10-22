package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/web/auth"
	"tuyou/internal/repository"
	"tuyou/internal/service/web/auth"
	"tuyou/internal/service/web/auth/sms"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"
	"tuyou/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authSvc     auth.WebAuthService
	redisRepo   repository.RedisRepository
	redisClient *redis.Client
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authSvc auth.WebAuthService, redisRepo repository.RedisRepository, redisClient *redis.Client) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, redisRepo: redisRepo, redisClient: redisClient}
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

// GetCaptcha 获取图形验证码
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	resp, err := h.authSvc.GetCaptcha(c.Request.Context())
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

// Refresh 使用刷新令牌换取新的访问令牌与刷新令牌
// POST /api/v1/auth/refresh（公开接口，仅凭刷新令牌调用）
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken)
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

	// 改绑手机号/注销发码需登录；register/login/reset-pwd 三个场景保持公开
	if req.Scene == string(sms.SceneChangePhone) || req.Scene == string(sms.SceneCancelAccount) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if _, err := jwt.ParseToken(token); err != nil {
			response.Unauthorized(c, "未登录")
			return
		}
	}
	if err := sms.ValidatorPhone(req.Phone); err != nil {
		response.BizError(c, err)
		return
	}
	if err := reserveSMS(c.Request.Context(), h.redisClient, req.Phone); err != nil {
		if _, ok := err.(*appErrors.BizError); ok {
			response.BizError(c, err)
		} else {
			response.InternalError(c, "短信服务暂不可用")
		}
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

// ChangePhone 修改绑定手机号
// POST /api/v1/web/auth/change-phone（需登录）
func (h *AuthHandler) ChangePhone(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req request.ChangePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authSvc.ChangePhone(c.Request.Context(), userID, &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	// 成功响应：message 取服务层文案，data 为空
	response.SuccessWithMessage(c, resp.Message, struct{}{})
}

// UpdatePassword 修改密码
// PUT /api/v1/web/auth/password（需登录）
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req request.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		response.BadRequest(c, "两次输入密码不一致")
		return
	}

	resp, err := h.authSvc.UpdatePassword(c.Request.Context(), userID, &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	// 成功响应：message 取服务层文案，data 为空
	response.SuccessWithMessage(c, resp.Message, struct{}{})
}

// Cancel 注销账号
// POST /api/v1/web/auth/cancel（需登录）
func (h *AuthHandler) Cancel(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}

	var req request.CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authSvc.Cancel(c.Request.Context(), userID, &req)
	if err != nil {
		response.BizError(c, err)
		return
	}
	// 成功响应：message 取服务层文案，data 为空
	response.SuccessWithMessage(c, resp.Message, struct{}{})
}

// Logout 拉黑当前 jti 使其立即失效；幂等，token 已失效时也返回成功
func (h *AuthHandler) Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if claims, err := jwt.ParseToken(token); err == nil {
		_ = h.authSvc.Logout(c.Request.Context(), claims.GetUserID(), claims.ID)
	}
	response.SuccessWithMessage(c, "退出登录成功", struct{}{})
}
