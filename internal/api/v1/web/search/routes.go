package search

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册全量搜索路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	// 热门搜索关键词（需要登录）
	hot := r.Group("/hot-searches")
	hot.Use(middleware.MemberAuth(c.redisRepo))
	{
		hot.GET("", c.GetHotSearches)
	}
}
