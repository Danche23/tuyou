package auth

import "github.com/gin-gonic/gin"

func (h *AdminHandler) RegisterRoutes(group *gin.RouterGroup) {
	adminGroup := group.Group("/admin")
	{
		adminGroup.GET("/captcha", h.GetCaptcha)
		adminGroup.POST("/login", h.Login)
	}
}
