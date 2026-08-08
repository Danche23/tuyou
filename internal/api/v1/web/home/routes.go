package home

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册首页路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	// 首页轮播（需要登录）
	slides := r.Group("/slides")
	slides.Use(middleware.MemberAuth(c.redisRepo))
	{
		slides.GET("", c.GetSlides)
	}

	// Feed 流（需要登录，瀑布流分页）
	feed := r.Group("/feed")
	feed.Use(middleware.MemberAuth(c.redisRepo))
	{
		feed.GET("", c.GetFeed)
	}
}
