package workbench

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台工作台路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin/workbench")
	admin.Use(middleware.Auth(c.redisRepo))
	{
		admin.GET("", c.GetWorkbench)
	}
}
