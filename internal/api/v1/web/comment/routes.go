package comment

// 前台评论模块 —— 路由注册
//
// 5 个评论接口全部挂在 /api/v1/comments 下，且整组套 middleware.MemberAuth：
// 登录中间件（校验 JWT + 白名单），所以每个 Handler 里拿到的 memberID 一定是真实登录用户。
//
// 路由分层：一级评论是 /comments、/comments/:id；
// 点赞和二级评论为了避免和 /comments/:id 冲突，单独放在
// /comments/like/:id、/comments/replies/:id 两个前缀下。
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
