package message

import (
	"strconv"

	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/admin/message"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/message"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 后台消息管理控制器
type Controller struct {
	msgSvc    service.MessageService
	redisRepo repository.RedisRepository
}

// NewController 创建控制器实例
func NewController(msgSvc service.MessageService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{msgSvc: msgSvc, redisRepo: redisRepo}
}

// ListInteractions 分页查询互动消息
// GET /api/v1/admin/messages/interactions
func (c *Controller) ListInteractions(ctx *gin.Context) {
	var req request.ListMessageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.msgSvc.ListInteractions(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// ListSystem 分页查询系统消息
// GET /api/v1/admin/messages/system
func (c *Controller) ListSystem(ctx *gin.Context) {
	var req request.ListMessageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.msgSvc.ListSystem(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// PublishSystemMsg 发布系统消息
// POST /api/v1/admin/messages/system
func (c *Controller) PublishSystemMsg(ctx *gin.Context) {
	var req request.PublishSystemMsgReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	if err := c.msgSvc.PublishSystemMsg(ctx, middleware.GetUserID(ctx), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// Delete 硬删除消息
// DELETE /api/v1/admin/messages/:id
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(ctx, "消息ID不合法")
		return
	}
	if err := c.msgSvc.Delete(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
