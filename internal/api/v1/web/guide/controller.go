package guide

import (
	"net/http"
	"strconv"

	"tuyou/internal/middleware"
	requestGuide "tuyou/internal/model/dto/request/web/guide"
	serviceGuide "tuyou/internal/service/web/guide"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 攻略控制器
type Controller struct {
	guideSvc serviceGuide.GuideService
}

// NewController 创建控制器实例
func NewController(guideSvc serviceGuide.GuideService) *Controller {
	return &Controller{guideSvc: guideSvc}
}

// Create 发布攻略
// POST /api/v1/guides
func (c *Controller) Create(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestGuide.CreateGuideRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := c.guideSvc.CreateGuide(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetGuideDetail 攻略详情
// GET /api/v1/guides/:id
func (c *Controller) GetGuideDetail(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	strategyID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	resp, err := c.guideSvc.GetGuideDetail(uint(strategyID), memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// ToggleLike 点赞/取消点赞攻略
// POST /api/v1/guides/:id/like
func (c *Controller) ToggleLike(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	strategyID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	resp, err := c.guideSvc.ToggleLikeGuide(uint(strategyID), memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// ToggleCollect 收藏/取消收藏攻略
// POST /api/v1/guides/:id/favorite
func (c *Controller) ToggleCollect(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	strategyID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	resp, err := c.guideSvc.ToggleCollectGuide(uint(strategyID), memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
