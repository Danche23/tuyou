package note

import (
	"net/http"

	"tuyou/internal/middleware"
	requestNote "tuyou/internal/model/dto/request/web/note"
	serviceNote "tuyou/internal/service/web/note"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 笔记控制器
type Controller struct {
	noteSvc serviceNote.NoteService
}

// NewController 创建控制器实例
func NewController(noteSvc serviceNote.NoteService) *Controller {
	return &Controller{noteSvc: noteSvc}
}

// Create 发布笔记
// POST /api/v1/notes
func (c *Controller) Create(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestNote.CreateNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := c.noteSvc.Create(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
