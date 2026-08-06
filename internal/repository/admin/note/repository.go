package note

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

type noteRepository struct {
	db *gorm.DB
}

// NewNoteRepository 创建后台笔记仓储实例
func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
}

// applyFilters 拼装动态查询条件（针对 note 表）
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["title"]; ok {
		query = query.Where("title LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["category"]; ok {
		query = query.Where("category = ?", v)
	}
	if v, ok := filters["is_slide"]; ok {
		query = query.Where("is_slide = ?", v)
	}
	if v, ok := filters["begin_time"]; ok {
		query = query.Where("publish_time >= ?", v)
	}
	if v, ok := filters["finish_time"]; ok {
		query = query.Where("publish_time <= ?", v)
	}
	return query
}

// Count 统计笔记数量
func (r *noteRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Note{})
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindList 分页查询笔记列表
func (r *noteRepository) FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Note, error) {
	notes := make([]entity.Note, 0)

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Note{})
	query = applyFilters(query, filters)

	// 排序（sortField 来自 Service 白名单，安全）
	if sortField != "" {
		query = query.Order(sortField + " " + sortOrder)
	} else {
		query = query.Order("id DESC")
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

// FindTagsByNoteIDs 批量查询笔记标签 map[noteID][]tagName
func (r *noteRepository) FindTagsByNoteIDs(ctx context.Context, noteIDs []uint) (map[uint][]string, error) {
	if len(noteIDs) == 0 {
		return map[uint][]string{}, nil
	}

	var rows []struct {
		NoteID  uint   `gorm:"column:note_id"`
		TagName string `gorm:"column:tag_name"`
	}
	if err := r.db.WithContext(ctx).Table("note_tag").
		Select("note_tag.note_id AS note_id, tag.tag_name AS tag_name").
		Joins("JOIN tag ON tag.id = note_tag.tag_id").
		Where("note_tag.note_id IN ?", noteIDs).
		Order("note_tag.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	tagsMap := make(map[uint][]string)
	for _, row := range rows {
		tagsMap[row.NoteID] = append(tagsMap[row.NoteID], row.TagName)
	}
	return tagsMap, nil
}

// FindMembersByIDs 批量查询作者信息
func (r *noteRepository) FindMembersByIDs(ctx context.Context, memberIDs []uint) ([]entity.Member, error) {
	if len(memberIDs) == 0 {
		return []entity.Member{}, nil
	}

	members := make([]entity.Member, 0)
	if err := r.db.WithContext(ctx).Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// FindCoverImagesByNoteIDs 批量查询笔记封面图（每篇第一张）map[noteID]url
func (r *noteRepository) FindCoverImagesByNoteIDs(ctx context.Context, noteIDs []uint) (map[uint]string, error) {
	if len(noteIDs) == 0 {
		return map[uint]string{}, nil
	}

	var rows []struct {
		NoteID uint   `gorm:"column:note_id"`
		URL    string `gorm:"column:url"`
	}
	if err := r.db.WithContext(ctx).Table("note_image").
		Select("note_id AS note_id, url AS url").
		Where("note_id IN ?", noteIDs).
		Order("sort ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 行已按 sort 升序排列，每篇笔记取第一次出现的那张（即 sort 最小）
	covers := make(map[uint]string)
	seen := make(map[uint]bool)
	for _, row := range rows {
		if seen[row.NoteID] {
			continue
		}
		seen[row.NoteID] = true
		covers[row.NoteID] = row.URL
	}
	return covers, nil
}

// FindByID 根据 ID 查询单篇笔记（Unscoped 含软删除记录，不存在返回 nil, nil）
func (r *noteRepository) FindByID(ctx context.Context, id uint) (*entity.Note, error) {
	var note entity.Note
	err := r.db.WithContext(ctx).Unscoped().First(&note, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// Update 更新笔记字段（updates map 中的 key 都会被更新；Unscoped 以支持恢复已软删笔记）
func (r *noteRepository) Update(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Unscoped().Model(&entity.Note{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateCommentStatusByTarget 更新目标笔记/攻略下的评论状态
func (r *noteRepository) UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error {
	return tx.Model(&entity.Comment{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("status", status).Error
}

// IncrPublishCount 作者发布数 +1
func (r *noteRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}

// DecrPublishCount 作者发布数 -1
func (r *noteRepository) DecrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
}

// Delete 硬删除笔记及其关联数据（事务）；decrPublish=true 时作者发布数 -1
func (r *noteRepository) Delete(ctx context.Context, noteID uint, memberID uint, decrPublish bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 关联数据：图片 / 标签 / 点赞 / 收藏
		if err := tx.Unscoped().Where("note_id = ?", noteID).Delete(&entity.NoteImage{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("note_id = ?", noteID).Delete(&entity.NoteTag{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("note_id = ?", noteID).Delete(&entity.NoteLike{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("note_id = ?", noteID).Delete(&entity.NoteCollect{}).Error; err != nil {
			return err
		}
		// 评论点赞（先删，子查询引用 comment 表）
		if err := tx.Unscoped().
			Where("comment_id IN (SELECT id FROM comment WHERE target_type = ? AND target_id = ?)", 1, noteID).
			Delete(&entity.CommentLike{}).Error; err != nil {
			return err
		}
		// 评论（笔记 target_type=1）
		if err := tx.Unscoped().Where("target_type = ? AND target_id = ?", 1, noteID).Delete(&entity.Comment{}).Error; err != nil {
			return err
		}
		// 主行
		if err := tx.Unscoped().Delete(&entity.Note{}, noteID).Error; err != nil {
			return err
		}
		// 仍上架的笔记：作者发布数 -1
		if decrPublish {
			return tx.Model(&entity.Member{}).Where("id = ?", memberID).
				UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
		}
		return nil
	})
}
