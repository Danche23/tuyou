package home

import (
	"strconv"
	"strings"

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
		Joins("JOIN tag ON tag.id = note_tag.tag_id AND tag.status = 1 AND tag.delete_time IS NULL").
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
		Joins("JOIN tag ON tag.id = strategy_tag.tag_id AND tag.status = 1 AND tag.delete_time IS NULL").
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

// CountFeedNotes 统计符合条件的笔记总数（仅上架）
func (r *homeRepository) CountFeedNotes(category int) (int64, error) {
	var count int64
	q := r.db.Model(&entity.Note{}).Where("status = ?", 1)
	if category > 0 {
		q = q.Where("category = ?", category)
	}
	err := q.Count(&count).Error
	return count, err
}

// CountFeedStrategies 统计符合条件的攻略总数（仅上架）
func (r *homeRepository) CountFeedStrategies(category int) (int64, error) {
	var count int64
	q := r.db.Model(&entity.Strategy{}).Where("status = ?", 1)
	if category > 0 {
		q = q.Where("category = ?", category)
	}
	err := q.Count(&count).Error
	return count, err
}

// FindHotTagIDs 查询热门标签ID（is_hot=1 且启用未删）
func (r *homeRepository) FindHotTagIDs() ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&entity.Tag{}).
		Where("is_hot = ? AND status = ?", 1, 1).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindHotNoteIDs 查询带热门标签的笔记ID集合（去重）
func (r *homeRepository) FindHotNoteIDs(hotTagIDs []uint) ([]uint, error) {
	if len(hotTagIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	if err := r.db.Model(&entity.NoteTag{}).
		Where("tag_id IN ?", hotTagIDs).
		Distinct("note_id").
		Pluck("note_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindHotStrategyIDs 查询带热门标签的攻略ID集合（去重）
func (r *homeRepository) FindHotStrategyIDs(hotTagIDs []uint) ([]uint, error) {
	if len(hotTagIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	if err := r.db.Model(&entity.StrategyTag{}).
		Where("tag_id IN ?", hotTagIDs).
		Distinct("strategy_id").
		Pluck("strategy_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindFeedNotes 按排序窗口取笔记（limit = offset+size，软删除由 GORM 自动过滤）
// hotIDs 非空时，ORDER BY 前拼接热门置顶 CASE，保证热门内容整体排在前面。
func (r *homeRepository) FindFeedNotes(category int, order string, limit int, hotIDs []uint) ([]entity.Note, error) {
	var list []entity.Note
	q := r.db.Where("status = ?", 1)
	if category > 0 {
		q = q.Where("category = ?", category)
	}
	if len(hotIDs) > 0 {
		order = "CASE WHEN id IN (" + joinUintIDs(hotIDs) + ") THEN 0 ELSE 1 END, " + order
	}
	err := q.Order(order).Limit(limit).Find(&list).Error
	return list, err
}

// FindFeedStrategies 按排序窗口取攻略（limit = offset+size，软删除由 GORM 自动过滤）
// hotIDs 非空时，ORDER BY 前拼接热门置顶 CASE，保证热门内容整体排在前面。
func (r *homeRepository) FindFeedStrategies(category int, order string, limit int, hotIDs []uint) ([]entity.Strategy, error) {
	var list []entity.Strategy
	q := r.db.Where("status = ?", 1)
	if category > 0 {
		q = q.Where("category = ?", category)
	}
	if len(hotIDs) > 0 {
		order = "CASE WHEN id IN (" + joinUintIDs(hotIDs) + ") THEN 0 ELSE 1 END, " + order
	}
	err := q.Order(order).Limit(limit).Find(&list).Error
	return list, err
}

// joinUintIDs 把 uint 切片拼成逗号分隔的字符串，用于安全内联进 ORDER BY CASE（ID 来自数据库，安全）
func joinUintIDs(ids []uint) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatUint(uint64(id), 10)
	}
	return strings.Join(parts, ",")
}

// FindMembersByIDs 批量查询会员信息
func (r *homeRepository) FindMembersByIDs(memberIDs []uint) (map[uint]*entity.Member, error) {
	result := make(map[uint]*entity.Member)
	if len(memberIDs) == 0 {
		return result, nil
	}
	var members []entity.Member
	if err := r.db.Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	for i := range members {
		result[members[i].ID] = &members[i]
	}
	return result, nil
}

// FindNoteLikeStatuses 当前用户对各笔记的点赞状态（批量，供缓存未命中时兜底）
func (r *homeRepository) FindNoteLikeStatuses(noteIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if len(noteIDs) == 0 {
		return result, nil
	}
	var likes []entity.NoteLike
	if err := r.db.Where("note_id IN ? AND member_id = ?", noteIDs, memberID).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, l := range likes {
		result[l.NoteID] = true
	}
	return result, nil
}

// FindStrategyLikeStatuses 当前用户对各攻略的点赞状态（批量，供缓存未命中时兜底）
func (r *homeRepository) FindStrategyLikeStatuses(strategyIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if len(strategyIDs) == 0 {
		return result, nil
	}
	var likes []entity.StrategyLike
	if err := r.db.Where("strategy_id IN ? AND member_id = ?", strategyIDs, memberID).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, l := range likes {
		result[l.StrategyID] = true
	}
	return result, nil
}
