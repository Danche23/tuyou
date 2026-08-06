package note

import (
	"context"

	"gorm.io/gorm"

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

	// FindByID 根据 ID 查询单篇笔记，不存在时返回 nil, nil
	FindByID(ctx context.Context, id uint) (*entity.Note, error)

	// Update 更新笔记字段（updates map 中的 key 都会被更新；Unscoped，可更新软删行以支持恢复）
	Update(tx *gorm.DB, id uint, updates map[string]interface{}) error

	// UpdateCommentStatusByTarget 更新目标笔记/攻略下的评论状态
	UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error

	// IncrPublishCount 作者发布数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error

	// DecrPublishCount 作者发布数 -1
	DecrPublishCount(tx *gorm.DB, memberID uint) error

	// Delete 硬删除笔记及其关联数据（事务）；decrPublish=true 时作者发布数 -1
	Delete(ctx context.Context, noteID uint, memberID uint, decrPublish bool) error
}
