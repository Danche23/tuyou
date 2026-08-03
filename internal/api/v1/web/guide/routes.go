package guide

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册攻略路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	guides := r.Group("/guides")
	guides.Use(middleware.Auth())
	{
		guides.POST("", c.Create)
		guides.GET("/:id", c.GetGuideDetail)
		guides.POST("/:id/like", c.ToggleLike)
		guides.POST("/:id/favorite", c.ToggleCollect)
	}
}
