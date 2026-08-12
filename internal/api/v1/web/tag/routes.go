package tag

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册标签路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	common := r.Group("/common")
	common.Use(middleware.MemberAuth(c.redisRepo))
	{
		common.GET("/tags", c.ListTags)
	}
}
