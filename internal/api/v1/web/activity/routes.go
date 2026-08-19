package activity

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册活动路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	activities := r.Group("/activities")
	// 公开接口（无需登录）
	activities.GET("/:id", c.GetDetail)

	// 需要登录的接口
	activities.Use(middleware.MemberAuth(c.redisRepo))
	{
		activities.POST("", c.Create)
		activities.GET("", c.List)
		activities.POST("/:id/signup", c.ToggleSignup)
		activities.DELETE("/:id", c.Delete)
	}
}
