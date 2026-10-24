package message

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"tuyou/internal/stream"

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
	hub       *stream.SSEHub
}

// NewController 创建消息控制器实例
func NewController(msgSvc serviceMsg.MessageService, redisRepo repository.RedisRepository, hub *stream.SSEHub) *Controller {
	return &Controller{msgSvc: msgSvc, redisRepo: redisRepo, hub: hub}
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

// Delete 删除消息（软删除）
// DELETE /api/v1/messages/:id
func (c *Controller) Delete(ctx *gin.Context) {
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

	if err := c.msgSvc.Delete(memberID, uint(msgID)); err != nil {
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

func (c *Controller) Stream(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	// 注册连接，断开时自动移除
	client := c.hub.Subscribe(memberID)
	defer c.hub.Unsubscribe(memberID, client)

	// SSE 响应头：让浏览器/代理知道这是长连接流式响应
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("X-Accel-Buffering", "no")

	// http.Server 配置了 ReadTimeout / WriteTimeout（各 5 分钟），它们会给连接设置
	// 绝对读写截止时间。SSE 是长连接，不清除这两个 deadline 的话连接会在 5 分钟后
	// 被服务端强制断开（表现为前端每 5 分钟重连一次、期间事件可能丢失）。
	if rc := http.NewResponseController(ctx.Writer); rc != nil {
		_ = rc.SetReadDeadline(time.Time{})
		_ = rc.SetWriteDeadline(time.Time{})
	}

	ctx.Writer.Flush()

	//连接建立先推次当前未读数 ->前端兜底断线期间漏掉的消息
	if uc, err := c.msgSvc.UnreadCount(memberID); err == nil {
		c.writeEvent(ctx, "unread", uc)
	}

	//30秒心跳：注释行“：ping”，防止代理/负载均衡超时掐断空闲连接
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case data := <-client.Events:
			ctx.Writer.Write(data)
			ctx.Writer.Flush()
		case <-ticker.C:
			ctx.Writer.WriteString("：ping\n\n")
			ctx.Writer.Flush()
		}
	}
}

// writeEvent 写一帧SSE事件并立即flush
func (c *Controller) writeEvent(ctx *gin.Context, event string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	ctx.Writer.Write([]byte("event: " + event + "\ndata: " + string(payload) + "\n\n"))
	ctx.Writer.Flush()
}
