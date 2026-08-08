package home

import (
	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

type homeRepository struct {
	db *gorm.DB
}

func NewHomeRepository(db *gorm.DB) HomeRepository {
	return &homeRepository{db: db}
}

// ==================== 轮播：管理员已标记 is_slide=1（优先级 ①②③） ====================

// FindSlideActivities 轮播活动（排除已删除、已结束）
func (r *homeRepository) FindSlideActivities() ([]entity.Activity, error) {
	var list []entity.Activity
	err := r.db.Where("is_slide = ? AND is_del = ? AND activity_status != ?", 1, 0, 3).
		Order("create_time DESC").Limit(5).Find(&list).Error
	return list, err
}

// FindSlideNotes 轮播笔记（仅上架，软删除由 GORM 自动过滤）
func (r *homeRepository) FindSlideNotes() ([]entity.Note, error) {
	var list []entity.Note
	err := r.db.Where("is_slide = ? AND status = ?", 1, 1).
		Order("publish_time DESC").Limit(5).Find(&list).Error
	return list, err
}

// FindSlideStrategies 轮播攻略（仅上架）
func (r *homeRepository) FindSlideStrategies() ([]entity.Strategy, error) {
	var list []entity.Strategy
	err := r.db.Where("is_slide = ? AND status = ?", 1, 1).
		Order("publish_time DESC").Limit(5).Find(&list).Error
	return list, err
}

// ==================== 轮播：补齐候选 is_slide=0（优先级 ④⑤⑥，自动选中不写库） ====================

// FindFillActivities 补齐候选活动（is_slide=0，排除已删除、已结束）
func (r *homeRepository) FindFillActivities(limit int) ([]entity.Activity, error) {
	var list []entity.Activity
	err := r.db.Where("is_slide = ? AND is_del = ? AND activity_status != ?", 0, 0, 3).
		Order("create_time DESC").Limit(limit).Find(&list).Error
	return list, err
}

// FindFillNotes 补齐候选笔记（is_slide=0，仅上架）
func (r *homeRepository) FindFillNotes(limit int) ([]entity.Note, error) {
	var list []entity.Note
	err := r.db.Where("is_slide = ? AND status = ?", 0, 1).
		Order("publish_time DESC").Limit(limit).Find(&list).Error
	return list, err
}

// FindFillStrategies 补齐候选攻略（is_slide=0，仅上架）
func (r *homeRepository) FindFillStrategies(limit int) ([]entity.Strategy, error) {
	var list []entity.Strategy
	err := r.db.Where("is_slide = ? AND status = ?", 0, 1).
		Order("publish_time DESC").Limit(limit).Find(&list).Error
	return list, err
}

// ==================== 关联数据（批量，避免 N+1；活动数据在主表，无需查询） ====================

// FindFirstImagesByNoteIDs 每个笔记取第一张图（sort 最小）作为封面
func (r *homeRepository) FindFirstImagesByNoteIDs(noteIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(noteIDs) == 0 {
		return result, nil
	}
	var imgs []entity.NoteImage
	if err := r.db.Where("note_id IN ?", noteIDs).Order("sort ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	for _, img := range imgs {
		if _, ok := result[img.NoteID]; !ok {
			result[img.NoteID] = img.URL
		}
	}
	return result, nil
}

// FindFirstImagesByStrategyIDs 每个攻略取第一张图（sort 最小）作为封面
func (r *homeRepository) FindFirstImagesByStrategyIDs(strategyIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(strategyIDs) == 0 {
		return result, nil
	}
	var imgs []entity.StrategyImage
	if err := r.db.Where("strategy_id IN ?", strategyIDs).Order("sort ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	for _, img := range imgs {
		if _, ok := result[img.StrategyID]; !ok {
			result[img.StrategyID] = img.URL
		}
	}
	return result, nil
}

// FindFirstTagsByNoteIDs 每个笔记取第一个启用标签
func (r *homeRepository) FindFirstTagsByNoteIDs(noteIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(noteIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		NoteID  uint
		TagName string
	}
	if err := r.db.Table("note_tag").
		Select("note_tag.note_id, tag.tag_name").
		Joins("JOIN tag ON tag.id = note_tag.tag_id AND tag.status = 1").
		Where("note_tag.note_id IN ?", noteIDs).
		Order("note_tag.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if _, ok := result[row.NoteID]; !ok {
			result[row.NoteID] = row.TagName
		}
	}
	return result, nil
}

// FindFirstTagsByStrategyIDs 每个攻略取第一个启用标签
func (r *homeRepository) FindFirstTagsByStrategyIDs(strategyIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(strategyIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		StrategyID uint
		TagName    string
	}
	if err := r.db.Table("strategy_tag").
		Select("strategy_tag.strategy_id, tag.tag_name").
		Joins("JOIN tag ON tag.id = strategy_tag.tag_id AND tag.status = 1").
		Where("strategy_tag.strategy_id IN ?", strategyIDs).
		Order("strategy_tag.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if _, ok := result[row.StrategyID]; !ok {
			result[row.StrategyID] = row.TagName
		}
	}
	return result, nil
}
