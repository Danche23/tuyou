package activity

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台活动路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.POST("/activities", c.CreateActivity)
		admin.GET("/activities", c.ListActivities)
		admin.GET("/activities/:id", c.GetActivityDetail)
		admin.PUT("/activities/:id", c.UpdateActivity)
		admin.DELETE("/activities/:id", c.DeleteActivity)
		admin.POST("/activities/batch-delete", c.BatchDeleteActivities)
		admin.POST("/activities/export", c.ExportActivities)
		admin.GET("/leadernames", c.ListLeaderNames)
	}
}
