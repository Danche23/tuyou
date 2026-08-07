package guide

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/guide"
	service "tuyou/internal/service/admin/guide"
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

// parseGuideID 解析路径参数 id（兼容 "G0001" 展示格式与纯数字）
func parseGuideID(s string) (uint, error) {
	id, err := strconv.ParseUint(strings.TrimPrefix(s, "G"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
