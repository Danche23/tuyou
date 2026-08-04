package center

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册个人中心路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin/center")
	admin.Use(middleware.Auth())
	{
		admin.PUT("/avatar", c.UpdateAvatar)
		admin.PUT("/profile", c.UpdateProfile)
	}
}
