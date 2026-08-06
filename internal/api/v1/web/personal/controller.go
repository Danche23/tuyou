package personal

import (
	"net/http"

	"tuyou/internal/middleware"
	servicePersonal "tuyou/internal/service/web/personal"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 个人中心控制器
type Controller struct {
	personalSvc servicePersonal.PersonalService
}

// NewController 创建控制器实例
func NewController(personalSvc servicePersonal.PersonalService) *Controller {
	return &Controller{personalSvc: personalSvc}
}

// GetFollowedMembers 我关注的用户列表
// GET /api/v1/personal/following
func (c *Controller) GetFollowedMembers(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := c.personalSvc.GetFollowedMembers(memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

func (c *Controller) GetFansMembers(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := c.personalSvc.GetFansMembers(memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
