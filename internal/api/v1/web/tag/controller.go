package tag

import (
	"net/http"

	"tuyou/internal/middleware"
	serviceTag "tuyou/internal/service/web/tag"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 标签控制器
type Controller struct {
	tagSvc serviceTag.TagService
}

// NewController 创建控制器实例
func NewController(tagSvc serviceTag.TagService) *Controller {
	return &Controller{tagSvc: tagSvc}
}

// ListTags 获取启用标签列表
// GET /api/v1/common/tags
func (c *Controller) ListTags(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := c.tagSvc.ListTags(ctx)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
