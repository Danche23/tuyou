package note

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"tuyou/internal/repository"

	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/note"
	service "tuyou/internal/service/admin/note"
	"tuyou/pkg/errors"
	"tuyou/pkg/response"
)

// Controller 后台笔记管理控制器
type Controller struct {
	noteSvc   service.NoteService
	redisRepo repository.RedisRepository
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

// GetNoteDetail 获取笔记详情
// GET /api/v1/admin/notes/:id
func (c *Controller) GetNoteDetail(ctx *gin.Context) {
	id, err := parseNoteID(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	resp, err := c.noteSvc.GetNoteDetail(ctx, id)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// UpdateNote 编辑笔记
// PUT /api/v1/admin/notes/:id
func (c *Controller) UpdateNote(ctx *gin.Context) {
	id, err := parseNoteID(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	var req request.UpdateNoteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.noteSvc.UpdateNote(ctx, id, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// DeleteNote 删除笔记
// DELETE /api/v1/admin/notes/:id
func (c *Controller) DeleteNote(ctx *gin.Context) {
	id, err := parseNoteID(ctx.Param("id"))
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	if err := c.noteSvc.DeleteNote(ctx, id); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// BatchDeleteNotes 批量删除笔记
// POST /api/v1/admin/notes/batch-delete
func (c *Controller) BatchDeleteNotes(ctx *gin.Context) {
	var req request.BatchDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	if err := c.noteSvc.BatchDeleteNotes(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ExportNotes 批量导出笔记（xlsx 下载）
// POST /api/v1/admin/notes/export
func (c *Controller) ExportNotes(ctx *gin.Context) {
	var req request.ExportNoteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	list, err := c.noteSvc.ExportNotes(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	data, err := buildNoteExcel(list)
	if err != nil {
		response.BizError(ctx, errors.NewWithErr(errors.CodeInternalError, "导出失败", err))
		return
	}

	filename := "note_export_" + time.Now().Format("20060102") + ".xlsx"
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// parseNoteID 解析路径参数 id（兼容 "N0001" 展示格式与纯数字，与批量接口 parseNoteIDs 对齐）
func parseNoteID(s string) (uint, error) {
	id, err := strconv.ParseUint(strings.TrimPrefix(s, "N"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
