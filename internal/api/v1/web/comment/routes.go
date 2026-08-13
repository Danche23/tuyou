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
		comments.POST("", c.Create)                 // 发表评论
		comments.GET("", c.List)                    // 评论列表
		comments.DELETE("/:id", c.Delete)           // 删除评论
		comments.POST("/like/:id", c.ToggleLike)    // 点赞/取消点赞
		comments.GET("/replies/:id", c.ListReplies) // 二级评论列表
	}
}
