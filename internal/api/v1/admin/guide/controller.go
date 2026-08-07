package guide

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/guide"
	service "tuyou/internal/service/admin/guide"
	"tuyou/pkg/errors"
	"tuyou/pkg/response"
)

// Controller 后台攻略管理控制器
type Controller struct {
	guideSvc service.GuideService
}

// NewController 创建控制器实例
func NewController(guideSvc service.GuideService) *Controller {
	return &Controller{guideSvc: guideSvc}
}

// ListGuides 分页查询攻略
// GET /api/v1/admin/guides
func (c *Controller) ListGuides(ctx *gin.Context) {
	var req request.ListGuideReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.guideSvc.ListGuides(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetGuideDetail 获取攻略详情
// GET /api/v1/admin/guides/:id
func (c *Controller) GetGuideDetail(ctx *gin.Context) {
	id, err := parseGuideID(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	resp, err := c.guideSvc.GetGuideDetail(ctx, id)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// UpdateGuide 编辑攻略
// PUT /api/v1/admin/guides/:id
func (c *Controller) UpdateGuide(ctx *gin.Context) {
	id, err := parseGuideID(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	var req request.UpdateGuideReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.guideSvc.UpdateGuide(ctx, id, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// DeleteGuide 删除攻略
// DELETE /api/v1/admin/guides/:id
func (c *Controller) DeleteGuide(ctx *gin.Context) {
	id, err := parseGuideID(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	if err := c.guideSvc.DeleteGuide(ctx, id); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// BatchDeleteGuides 批量删除攻略
// POST /api/v1/admin/guides/batch-delete
func (c *Controller) BatchDeleteGuides(ctx *gin.Context) {
	var req request.BatchDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.guideSvc.BatchDeleteGuides(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ExportGuides 批量导出攻略（xlsx 下载）
// POST /api/v1/admin/guides/export
func (c *Controller) ExportGuides(ctx *gin.Context) {
	var req request.ExportGuideReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	list, err := c.guideSvc.ExportGuides(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	data, err := buildGuideExcel(list)
	if err != nil {
		response.BizError(ctx, errors.NewWithErr(errors.CodeInternalError, "导出失败", err))
		return
	}

	filename := "guide_export_" + time.Now().Format("20060102") + ".xlsx"
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// parseGuideID 解析路径参数 id（兼容 "G0001" 展示格式与纯数字）
func parseGuideID(s string) (uint, error) {
	id, err := strconv.ParseUint(strings.TrimPrefix(s, "G"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
