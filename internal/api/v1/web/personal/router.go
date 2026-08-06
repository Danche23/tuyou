package personal

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册个人中心路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	personal := r.Group("/personal")
	personal.Use(middleware.Auth())
	{
		personal.GET("/following", c.GetFollowedMembers)
		personal.GET("/fans", c.GetFansMembers)
	}
}
