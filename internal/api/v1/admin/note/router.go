package note

import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台笔记路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth())
	{
		admin.GET("/notes", c.ListNotes)
		admin.GET("/notes/:id", c.GetNoteDetail)
	}
}
