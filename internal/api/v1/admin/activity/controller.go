package activity

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/activity"
	service "tuyou/internal/service/admin/activity"
	"tuyou/pkg/errors"
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

// CreateActivity 创建活动
// POST /api/v1/admin/activities
func (c *Controller) CreateActivity(ctx *gin.Context) {
	var req request.CreateActivityReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.activitySvc.CreateActivity(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
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

// GetActivityDetail 获取活动详情
// GET /api/v1/admin/activities/:id
func (c *Controller) GetActivityDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.activitySvc.GetActivityDetail(ctx, uint(id))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// UpdateActivity 编辑活动（分类/轮播/状态三字段整体覆盖）
// PUT /api/v1/admin/activities/:id
func (c *Controller) UpdateActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	var req request.UpdateActivityReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.activitySvc.UpdateActivity(ctx, uint(id), &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// DeleteActivity 删除活动（硬删，连带报名记录）
// DELETE /api/v1/admin/activities/:id
func (c *Controller) DeleteActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.activitySvc.DeleteActivity(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// BatchDeleteActivities 批量删除活动（硬删，连带报名记录）
// POST /api/v1/admin/activities/batch-delete
func (c *Controller) BatchDeleteActivities(ctx *gin.Context) {
	var req request.BatchDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.activitySvc.BatchDeleteActivities(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ExportActivities 批量导出活动（xlsx 下载）
// POST /api/v1/admin/activities/export
func (c *Controller) ExportActivities(ctx *gin.Context) {
	var req request.ExportActivityReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	list, err := c.activitySvc.ExportActivities(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	data, err := buildActivityExcel(list)
	if err != nil {
		response.BizError(ctx, errors.NewWithErr(errors.CodeInternalError, "导出失败", err))
		return
	}

	filename := "activity_export_" + time.Now().Format("20060102") + ".xlsx"
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ListLeaderNames 获取领队名称列表
// GET /api/v1/admin/leadernames
func (c *Controller) ListLeaderNames(ctx *gin.Context) {
	resp, err := c.activitySvc.ListLeaderNames(ctx)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
