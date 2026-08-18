package auth

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) RegisterRoutes(group *gin.RouterGroup) {
	authGroup := group.Group("/auth")
	{
		authGroup.GET("/captcha", h.GetCaptcha)
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/reset-password", h.ResetPassword)
		authGroup.POST("/send-code", h.SendCode)
		authGroup.POST("/change-phone", middleware.MemberAuth(h.redisRepo), h.ChangePhone)
		authGroup.PUT("/password", middleware.MemberAuth(h.redisRepo), h.UpdatePassword)
	}
}
