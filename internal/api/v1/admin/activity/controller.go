package activity

import (
	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/activity"
	service "tuyou/internal/service/admin/activity"
	"tuyou/pkg/response"
)

// Controller 后台活动管理控制器
type Controller struct {
	activitySvc service.ActivityService
}

// NewController 创建控制器实例
func NewController(activitySvc service.ActivityService) *Controller {
	return &Controller{activitySvc: activitySvc}
}

// ListActivities 分页查询活动
// GET /api/v1/admin/activities
func (c *Controller) ListActivities(ctx *gin.Context) {
	var req request.ListActivityReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.activitySvc.ListActivities(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
