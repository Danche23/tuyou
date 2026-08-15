package auth

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (c *AdminController) RegisterRoutes(group *gin.RouterGroup) {
	adminGroup := group.Group("/admin")
	{
		adminGroup.GET("/captcha", c.GetCaptcha)
		adminGroup.POST("/login", c.Login)
		adminGroup.POST("/logout", c.Logout)
	}

	// 需要登录的接口（启用登录态白名单校验）
	authGroup := adminGroup.Group("/auth")
	authGroup.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		authGroup.GET("/me", c.Me)
		authGroup.PUT("/password", c.UpdatePassword)
	}
}
