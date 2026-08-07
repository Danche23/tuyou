package guide

import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台攻略路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth())
	{
		admin.GET("/guides", c.ListGuides)
		admin.GET("/guides/:id", c.GetGuideDetail)
		admin.PUT("/guides/:id", c.UpdateGuide)
	}
}
