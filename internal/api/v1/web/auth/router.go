package auth

import "github.com/gin-gonic/gin"

func (h *AuthHandler) RegisterRoutes(group *gin.RouterGroup) {
	authGroup := group.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/reset-password", h.ResetPassword)
		authGroup.POST("/send-code", h.SendCode)
	}
}
