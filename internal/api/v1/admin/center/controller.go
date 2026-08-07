package center

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/admin/center"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/center"
	"tuyou/pkg/response"
)

// Controller 个人中心控制器
type Controller struct {
	centerSvc service.CenterService
	redisRepo repository.RedisRepository
}

// NewController 创建控制器实例
func NewController(centerSvc service.CenterService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{centerSvc: centerSvc, redisRepo: redisRepo}
}

// UpdateAvatar 更新头像
func (c *Controller) UpdateAvatar(ctx *gin.Context) {
	var req request.UpdateAvatarRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.centerSvc.UpdateAvatar(ctx, middleware.GetUserID(ctx), req.Avatar)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "头像更新成功", resp)
}

// UpdateProfile 编辑账号
func (c *Controller) UpdateProfile(ctx *gin.Context) {
	var req request.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.centerSvc.UpdateUsername(ctx, middleware.GetUserID(ctx), req.Username)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "用户账号更新成功", resp)
}
