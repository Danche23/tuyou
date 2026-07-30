package note

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册笔记路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	notes := r.Group("/notes")
	notes.Use(middleware.Auth())
	{
		notes.POST("", c.Create)
		notes.GET("/:id", c.GetNoteDetail)
		notes.POST("/:id/like", c.ToggleLike)
	}
}
