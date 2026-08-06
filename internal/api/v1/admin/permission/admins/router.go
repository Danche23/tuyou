package admins

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
)

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(c.redisRepo))
	{
		admin.GET("/admins", c.List)
		admin.POST("/admins", c.Insert)
		admin.PUT("/admins/:id", c.Update)
		admin.DELETE("/admins/:id", c.Delete)
		admin.PUT("/admins/:id/reset-password", c.ResetPassword)
	}
}
