package tag

import (
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
}

// NewController 创建控制器实例
func NewController(tagSvc service.TagService) *Controller {
	return &Controller{tagSvc: tagSvc}
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
