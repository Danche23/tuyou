package message

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册消息路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	messages := r.Group("/messages")
	messages.Use(middleware.MemberAuth(c.redisRepo))
	{
		messages.GET("/interactions", c.GetInteractions)
		messages.GET("/system", c.GetSystem)
		messages.GET("/unread-count", c.GetUnreadCount)
		messages.GET("/stream", c.Stream) //SSE通道
		messages.PUT("/:id/read", c.MarkRead)
		messages.PUT("/read-all", c.MarkAllRead)
		messages.DELETE("/:id", c.Delete)
		messages.POST("/interactions/:id/reply", c.Reply)
	}
}
