package note

import (
	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/note"
	service "tuyou/internal/service/admin/note"
	"tuyou/pkg/response"
)

// Controller 后台笔记管理控制器
type Controller struct {
	noteSvc service.NoteService
}

// NewController 创建控制器实例
func NewController(noteSvc service.NoteService) *Controller {
	return &Controller{noteSvc: noteSvc}
}

// ListNotes 分页查询笔记
// GET /api/v1/admin/notes
func (c *Controller) ListNotes(ctx *gin.Context) {
	var req request.ListNoteReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.noteSvc.ListNotes(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
