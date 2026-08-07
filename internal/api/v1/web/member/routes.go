package member

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册会员路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	members := r.Group("/member")
	members.Use(middleware.MemberAuth(c.redisRepo))
	{
		members.GET("/info", c.GetBaseInfo)
	}
}
