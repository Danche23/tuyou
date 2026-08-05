package activity

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册活动路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	activities := r.Group("/activities")
	activities.Use(middleware.Auth())
	{
		activities.POST("", c.Create)
		activities.GET("", c.List)
		activities.GET("/:id", c.GetDetail)
		activities.POST("/:id/signup", c.ToggleSignup)
		activities.DELETE("/:id", c.Delete)
	}
}
