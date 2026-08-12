package tag

import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台标签路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(c.redisRepo))
	{
		admin.GET("/tags", c.ListTags)
	}
}
