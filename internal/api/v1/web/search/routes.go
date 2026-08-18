package search

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册全量搜索路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	// 全量搜索 / 热门搜索关键词（均需登录）
	searchGroup := r.Group("")
	searchGroup.Use(middleware.MemberAuth(c.redisRepo))
	{
		searchGroup.GET("/search", c.Search)
		searchGroup.GET("/hot-searches", c.GetHotSearches)
		searchGroup.GET("/search/history", c.ListSearchHistory)
		searchGroup.DELETE("/search/history/:id", c.DeleteSearchHistory)
		searchGroup.DELETE("/search/history", c.ClearSearchHistory)
	}
}
