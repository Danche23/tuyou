package center

import (
	"encoding/json"
	"gorm.io/gorm"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
	request "tuyou/internal/model/dto/request/admin/center"
	responseCenter "tuyou/internal/model/dto/response/admin/center"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/center"
	"tuyou/internal/stream"
	"tuyou/pkg/response"
)

// Controller 个人中心控制器
type Controller struct {
	centerSvc service.CenterService
	redisRepo repository.RedisRepository
	adminHub  *stream.SSEHub
	db        *gorm.DB
}

// NewController 创建控制器实例
func NewController(centerSvc service.CenterService, redisRepo repository.RedisRepository, adminHub *stream.SSEHub, db *gorm.DB) *Controller {
	return &Controller{centerSvc: centerSvc, redisRepo: redisRepo, adminHub: adminHub, db: db}
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

// GetNotifications 分页获取个人消息
// GET /admin/center/notifications?page=1&size=10
func (c *Controller) GetNotifications(ctx *gin.Context) {
	var req request.NotificationListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数错误")
		return
	}
	resp, err := c.centerSvc.GetNotifications(ctx.Request.Context(), middleware.GetUserID(ctx), req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "获取成功", resp)
}

// GetUnreadCount 获取未读消息数
// GET /admin/center/notifications/unread
func (c *Controller) GetUnreadCount(ctx *gin.Context) {
	resp, err := c.centerSvc.GetUnreadCount(ctx.Request.Context(), middleware.GetUserID(ctx))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "获取成功", resp)
}

// MarkRead 单条消息标记为已读
// PUT /admin/center/notifications/:id/read
func (c *Controller) MarkRead(ctx *gin.Context) {
	msgID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || msgID == 0 {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}
	if err := c.centerSvc.MarkRead(ctx.Request.Context(), middleware.GetUserID(ctx), uint(msgID)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "已标记为已读", nil)
}

// MarkAllRead 全部标记为已读
// PUT /admin/center/notifications/read-all
func (c *Controller) MarkAllRead(ctx *gin.Context) {
	if err := c.centerSvc.MarkAllRead(ctx.Request.Context(), middleware.GetUserID(ctx)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "已全部标记为已读", emptyPage())
}

// Delete 删除单条消息
// DELETE /admin/center/notifications/:id
func (c *Controller) Delete(ctx *gin.Context) {
	msgID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || msgID == 0 {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}
	if err := c.centerSvc.Delete(ctx.Request.Context(), middleware.GetUserID(ctx), uint(msgID)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "消息删除成功", nil)
}

// ClearAll 清空所有消息
// DELETE /admin/center/notifications
func (c *Controller) ClearAll(ctx *gin.Context) {
	if err := c.centerSvc.ClearAll(ctx.Request.Context(), middleware.GetUserID(ctx)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "已清空所有消息", emptyPage())
}

// Stream SSE 通道：仅推送未读消息数变化（角标实时更新）
// GET /admin/center/notifications/stream
func (c *Controller) Stream(ctx *gin.Context) {
	adminID := middleware.GetUserID(ctx)
	// 注册连接，断开时自动移除
	client := c.adminHub.Subscribe(adminID)
	defer c.adminHub.Unsubscribe(adminID, client)

	// SSE 响应头：让浏览器/代理知道这是长连接流式响应
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Writer.Flush()

	// 连接建立先推当前未读数，兜底断线期间漏掉的消息
	if uc, err := c.centerSvc.GetUnreadCount(ctx.Request.Context(), adminID); err == nil {
		c.writeEvent(ctx, "unread", uc)
	}

	// 30秒心跳：注释行"：ping"，防止代理/负载均衡超时掐断空闲连接
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

// emptyPage 空分页数据（标记全部已读/清空后返回给前端）
func emptyPage() *responseCenter.AdminMsgPage {
	return &responseCenter.AdminMsgPage{List: []responseCenter.AdminNotification{}, Page: 1, Size: 10}
}
