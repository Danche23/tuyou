package home

import (
	"net/http"

	"tuyou/internal/middleware"
	requestHome "tuyou/internal/model/dto/request/web/home"
	"tuyou/internal/repository"
	serviceHome "tuyou/internal/service/web/home"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 首页控制器
type Controller struct {
	homeSvc   serviceHome.HomeService
	redisRepo repository.RedisRepository
}

// NewController 创建控制器实例
func NewController(homeSvc serviceHome.HomeService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{homeSvc: homeSvc, redisRepo: redisRepo}
}

// GetSlides 获取首页轮播
// GET /api/v1/slides
func (c *Controller) GetSlides(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := c.homeSvc.GetSlides()
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetFeed 获取Feed流
// GET /api/v1/feed?page=1&size=10&category=0&sort=推荐
func (c *Controller) GetFeed(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestHome.FeedRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	resp, err := c.homeSvc.GetFeed(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
