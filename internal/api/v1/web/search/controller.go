package search

import (
	"net/http"

	"tuyou/internal/middleware"
	requestSearch "tuyou/internal/model/dto/request/web/search"
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

// Search 全量搜索
// GET /api/v1/search?keyword=xx&type=content&page=1&size=10
// type=content（默认）：返回 FeedItem 分页列表（笔记+攻略）；type=activity：返回 ActivityListItem 分页列表。
func (c *Controller) Search(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestSearch.SearchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数不合法")
		return
	}

	list, total, err := c.searchSvc.Search(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(list, total, req.Page, req.Size))
}
