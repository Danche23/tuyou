package activity

import (
	"net/http"
	"strconv"

	"tuyou/internal/middleware"
	requestActivity "tuyou/internal/model/dto/request/web/activity"
	serviceActivity "tuyou/internal/service/web/activity"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 活动控制器
type Controller struct {
	activitySvc serviceActivity.ActivityService
}

// NewController 创建控制器实例
func NewController(activitySvc serviceActivity.ActivityService) *Controller {
	return &Controller{activitySvc: activitySvc}
}

// Create 创建活动
// POST /api/v1/activities
func (c *Controller) Create(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestActivity.CreateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := c.activitySvc.CreateActivity(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetDetail 获取活动详情
// GET /api/v1/activities/:id
func (c *Controller) GetDetail(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	resp, err := c.activitySvc.GetActivityDetail(uint(activityID), memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
