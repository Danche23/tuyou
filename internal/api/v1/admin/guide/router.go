package guide

import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台攻略路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.GET("/guides", c.ListGuides)
		admin.GET("/guides/:id", c.GetGuideDetail)
		admin.PUT("/guides/:id", c.UpdateGuide)
		admin.DELETE("/guides/:id", c.DeleteGuide)
		admin.POST("/guides/batch-delete", c.BatchDeleteGuides)
		admin.POST("/guides/export", c.ExportGuides)
	}
}
