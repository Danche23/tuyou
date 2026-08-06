package note

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/note"
	response "tuyou/internal/model/dto/response/admin/note"
)

// NoteService 后台笔记管理服务接口
type NoteService interface {
	// ListNotes 分页查询笔记列表
	ListNotes(ctx context.Context, req *request.ListNoteReq) (*response.ListNoteResp, error)

	// GetNoteDetail 获取笔记详情
	GetNoteDetail(ctx context.Context, id uint) (*response.NoteDetailResp, error)

	// UpdateNote 编辑笔记（分类/轮播/状态三字段必传，整体覆盖），返回更新后的详情
	UpdateNote(ctx context.Context, id uint, req *request.UpdateNoteReq) (*response.NoteDetailResp, error)

	// DeleteNote 硬删除笔记（连带关联数据；仍上架的笔记作者发布数 -1）
	DeleteNote(ctx context.Context, id uint) error

	// BatchDeleteNotes 批量硬删除笔记（连带关联数据；不存在的 ID 自动跳过）
	BatchDeleteNotes(ctx context.Context, req *request.BatchDeleteReq) error

	// ExportNotes 按 ID 批量导出笔记
	ExportNotes(ctx context.Context, req *request.ExportNoteReq) ([]response.NoteListItem, error)
}
