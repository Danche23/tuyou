package auth

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *AdminHandler) RegisterRoutes(group *gin.RouterGroup) {
	adminGroup := group.Group("/admin")
	{
		adminGroup.GET("/captcha", h.GetCaptcha)
		adminGroup.POST("/login", h.Login)
	}

	// 需要登录的接口
	authGroup := adminGroup.Group("/auth")
	authGroup.Use(middleware.Auth())
	{
		authGroup.GET("/me", h.Me)
		authGroup.PUT("/password", h.UpdatePassword)
	}
}
