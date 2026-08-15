package note

import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台笔记路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.GET("/notes", c.ListNotes)
		admin.GET("/notes/:id", c.GetNoteDetail)
		admin.PUT("/notes/:id", c.UpdateNote)
		admin.DELETE("/notes/:id", c.DeleteNote)
		admin.POST("/notes/batch-delete", c.BatchDeleteNotes)
		admin.POST("/notes/export", c.ExportNotes)
	}
}
