package sensitiveword

import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台敏感词管理路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.GET("/sensitive-words", c.List)
		admin.POST("/sensitive-words", c.Create)
		admin.DELETE("/sensitive-words/:id", c.Delete)
	}
}
