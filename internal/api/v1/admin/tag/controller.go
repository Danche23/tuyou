package tag

import (
	"gorm.io/gorm"
	"strconv"

	request "tuyou/internal/model/dto/request/admin/tag"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/tag"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 后台标签管理控制器
type Controller struct {
	tagSvc    service.TagService
	redisRepo repository.RedisRepository
	db        *gorm.DB
}

// NewController 创建控制器实例
func NewController(tagSvc service.TagService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{tagSvc: tagSvc, redisRepo: redisRepo, db: db}
}

// ListTags 分页查询标签
// GET /api/v1/admin/tags
func (c *Controller) ListTags(ctx *gin.Context) {
	var req request.ListTagReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.tagSvc.ListTags(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// CreateTag 新增标签
// POST /api/v1/admin/tags
func (c *Controller) CreateTag(ctx *gin.Context) {
	var req request.CreateTagReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.tagSvc.CreateTag(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// UpdateTag 编辑标签（状态/热门 必填；标签名称不可修改）
// PUT /api/v1/admin/tags/:id
func (c *Controller) UpdateTag(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "标签ID不合法")
		return
	}

	var req request.UpdateTagReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.tagSvc.UpdateTag(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// DeleteTag 删除标签（硬删，连带清理关联）
// DELETE /api/v1/admin/tags/:id
func (c *Controller) DeleteTag(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "标签ID不合法")
		return
	}

	if err := c.tagSvc.DeleteTag(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// BatchDeleteTags 批量删除标签（硬删，连带清理关联）
// POST /api/v1/admin/tags/batch-delete  body: {"ids":["1","2"]}
func (c *Controller) BatchDeleteTags(ctx *gin.Context) {
	var req request.BatchDeleteTagReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.tagSvc.BatchDeleteTags(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
