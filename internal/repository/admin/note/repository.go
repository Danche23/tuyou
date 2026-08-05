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
