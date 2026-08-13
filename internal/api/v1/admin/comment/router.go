package comment

// 后台评论管理模块 —— 路由注册
//
// 4 个接口挂在 /api/v1/admin/comments 下，整组套 middleware.Auth：
// 管理员鉴权中间件（校验 JWT + Redis 白名单）。
// 注意这里把 c.redisRepo 传了进去（白名单真正生效）——
// 对比 note/tag 后台模块传 nil（退化成纯 JWT 校验），评论模块是更严的写法。
//
// 接口清单：
//   GET    /admin/comments            评论分页列表（多条件筛选 + 排序）
//   GET    /admin/comments/:id        评论详情
//   DELETE /admin/comments/:id        彻底删除单条（一级连带其下所有二级）
//   POST   /admin/comments/batch-delete  批量彻底删除（body: {"ids":[...]}）
//
// 全部走"物理硬删"：评论行真的从表里删掉（前台软删只藏起来，见 web/comment/routes.go）。
import (
	"github.com/gin-gonic/gin"

	"tuyou/internal/middleware"
)

// RegisterRoutes 注册后台评论管理路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(c.redisRepo))
	{
		admin.GET("/comments", c.ListComments)
		admin.GET("/comments/:id", c.GetCommentDetail)
		admin.DELETE("/comments/:id", c.DeleteComment)
		admin.POST("/comments/batch-delete", c.BatchDeleteComments)
	}
}
