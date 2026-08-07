package member

import (
	"tuyou/internal/middleware"
	serviceAuth "tuyou/internal/service/web/auth"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 会员控制器
type Controller struct {
	authSvc serviceAuth.WebAuthService
}

// NewController 创建会员控制器
func NewController(authSvc serviceAuth.WebAuthService) *Controller {
	return &Controller{authSvc: authSvc}
}

// GetBaseInfo 获取当前登录会员基础信息
// GET /api/v1/web/member/info
func (c *Controller) GetBaseInfo(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "未登录")
		return
	}

	resp, err := c.authSvc.GetBaseInfo(ctx.Request.Context(), userID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
