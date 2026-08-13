package search

import (
	"net/http"

	"tuyou/internal/middleware"
	"tuyou/internal/repository"
	serviceSearch "tuyou/internal/service/web/search"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 全量搜索控制器
type Controller struct {
	searchSvc serviceSearch.SearchService
	redisRepo repository.RedisRepository
}

// NewController 创建控制器实例
func NewController(searchSvc serviceSearch.SearchService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{searchSvc: searchSvc, redisRepo: redisRepo}
}

// GetHotSearches 获取热门搜索关键词
// GET /api/v1/hot-searches
func (c *Controller) GetHotSearches(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	names, err := c.searchSvc.GetHotSearches()
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, names)
}
