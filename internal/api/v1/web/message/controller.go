package message

import (
	"net/http"
	"strconv"

	"tuyou/internal/middleware"
	requestMsg "tuyou/internal/model/dto/request/web/message"
	"tuyou/internal/repository"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 消息控制器
type Controller struct {
	msgSvc    serviceMsg.MessageService
	redisRepo repository.RedisRepository
}

// NewController 创建消息控制器实例
func NewController(msgSvc serviceMsg.MessageService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{msgSvc: msgSvc, redisRepo: redisRepo}
}

// GetInteractions 获取互动消息列表
// GET /api/v1/messages/interactions?page=1&size=10
func (c *Controller) GetInteractions(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestMsg.MessageListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	resp, err := c.msgSvc.GetInteractions(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetSystem 获取系统通知列表
// GET /api/v1/messages/system?page=1&size=10
func (c *Controller) GetSystem(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestMsg.MessageListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	resp, err := c.msgSvc.GetSystem(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// MarkRead 标记单条消息已读
// PUT /api/v1/messages/:id/read
func (c *Controller) MarkRead(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	msgID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || msgID == 0 {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	if err := c.msgSvc.MarkRead(memberID, uint(msgID)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// MarkAllRead 标记全部消息已读
// PUT /api/v1/messages/read-all
func (c *Controller) MarkAllRead(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	if err := c.msgSvc.MarkAllRead(memberID); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// Reply 消息回复评论
// POST /api/v1/messages/interactions/:id/reply
func (c *Controller) Reply(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	msgID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || msgID == 0 {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	var req requestMsg.ReplyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}

	if err := c.msgSvc.Reply(memberID, uint(msgID), req.Content); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// GetUnreadCount 未读消息数
// GET /api/v1/messages/unread-count
func (c *Controller) GetUnreadCount(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	resp, err := c.msgSvc.UnreadCount(memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
