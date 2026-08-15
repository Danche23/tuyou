package message

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册后台消息管理路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin/messages")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.GET("/interactions", c.ListInteractions)
		admin.GET("/system", c.ListSystem)
		admin.POST("/system", c.PublishSystemMsg)
		admin.DELETE("/:id", c.Delete)
	}
}
