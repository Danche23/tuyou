package note

import (
	"context"

	"tuyou/internal/model/entity"
)

// NoteRepository 后台笔记仓储接口
type NoteRepository interface {
	// Count 统计笔记数量（含筛选条件）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// FindList 分页查询笔记列表（含筛选条件、排序），返回实体
	FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Note, error)

	// FindTagsByNoteIDs 批量查询笔记标签 map[noteID][]tagName
	FindTagsByNoteIDs(ctx context.Context, noteIDs []uint) (map[uint][]string, error)

	// FindMembersByIDs 批量查询作者信息
	FindMembersByIDs(ctx context.Context, memberIDs []uint) ([]entity.Member, error)

	// FindCoverImagesByNoteIDs 批量查询笔记封面图（每篇第一张）map[noteID]url
	FindCoverImagesByNoteIDs(ctx context.Context, noteIDs []uint) (map[uint]string, error)
}
