package member

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册会员/领队管理路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(c.redisRepo))
	{
		// 会员
		admin.GET("/members", c.ListMembers)
		admin.GET("/members/:id", c.GetDetail)
		admin.PUT("/members/:id", c.Update)
		admin.DELETE("/members/:id", c.Delete)
		admin.POST("/members/batch-delete", c.BatchDelete)
		admin.PUT("/members/batch-status", c.BatchUpdateStatus)
		admin.GET("/members/export", c.ExportMembers)

		// 领队（同一套 handler，role 由 ListLeaders 固定传 0）
		admin.GET("/leaders", c.ListLeaders)
		admin.GET("/leaders/:id", c.GetDetail)
		admin.PUT("/leaders/:id", c.Update)
		admin.DELETE("/leaders/:id", c.Delete)
		admin.POST("/leaders/batch-delete", c.BatchDelete)
		admin.PUT("/leaders/batch-status", c.BatchUpdateStatus)
		admin.GET("/leaders/export", c.ExportLeaders)
	}
}
