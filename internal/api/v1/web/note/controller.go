package note

import (
	"net/http"
	"strconv"

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

	resp, err := c.noteSvc.CreateNote(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetNoteDetail 笔记详情
// GET /api/v1/notes/:id
func (c *Controller) GetNoteDetail(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	noteID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	resp, err := c.noteSvc.GetNoteDetail(uint(noteID), memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
