package center

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册个人中心路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin/center")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.PUT("/avatar", c.UpdateAvatar)
		admin.PUT("/profile", c.UpdateProfile)
		admin.GET("/notifications", c.GetNotifications)
		admin.GET("/notifications/unread", c.GetUnreadCount)
		admin.GET("/notifications/stream", c.Stream)
		admin.PUT("/notifications/:id/read", c.MarkRead)
		admin.PUT("/notifications/read-all", c.MarkAllRead)
		admin.DELETE("/notifications/:id", c.Delete)
		admin.DELETE("/notifications", c.ClearAll)
	}
}
