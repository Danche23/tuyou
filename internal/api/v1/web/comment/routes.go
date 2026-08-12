package comment

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册评论路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	comments := r.Group("/comments")
	comments.Use(middleware.MemberAuth(c.redisRepo))
	{
		comments.POST("/:id/like", c.ToggleLike)
	}
}
