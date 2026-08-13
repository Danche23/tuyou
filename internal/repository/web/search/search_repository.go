package search

import (
	"strings"

	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

type searchRepository struct {
	db *gorm.DB
}

// NewSearchRepository 创建搜索仓储实例
func NewSearchRepository(db *gorm.DB) SearchRepository {
	return &searchRepository{db: db}
}

// FindHotSearchNames 查询热门搜索关键词（status=1 AND is_hot=1 且未删除），按 id 升序
// 软删除由 GORM 自动过滤（delete_time IS NULL）。
func (r *searchRepository) FindHotSearchNames() ([]string, error) {
	var names []string
	if err := r.db.Model(&entity.Tag{}).
		Where("status = ? AND is_hot = ?", 1, 1).
		Order("id ASC").
		Pluck("tag_name", &names).Error; err != nil {
		return nil, err
	}
	return names, nil
}

// FindTagIDsByKeyword 关键词命中标签名（status=1 且未删除），返回标签ID集合
func (r *searchRepository) FindTagIDsByKeyword(keyword string) ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&entity.Tag{}).
		Where("status = ? AND tag_name LIKE ?", 1, "%"+escapeLike(keyword)+"%").
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindNoteIDsByTagIDs 带指定标签的笔记ID集合（去重）
func (r *searchRepository) FindNoteIDsByTagIDs(tagIDs []uint) ([]uint, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	if err := r.db.Model(&entity.NoteTag{}).
		Where("tag_id IN ?", tagIDs).
		Distinct("note_id").
		Pluck("note_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindStrategyIDsByTagIDs 带指定标签的攻略ID集合（去重）
func (r *searchRepository) FindStrategyIDsByTagIDs(tagIDs []uint) ([]uint, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	if err := r.db.Model(&entity.StrategyTag{}).
		Where("tag_id IN ?", tagIDs).
		Distinct("strategy_id").
		Pluck("strategy_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// SearchNotes 搜索笔记（标题/分类/标签 OR 命中），按热度倒序，取 limit 条（limit=offset+size）
func (r *searchRepository) SearchNotes(keyword string, category int, contentIDs []uint, limit int) ([]entity.Note, error) {
	var list []entity.Note
	cond, args := buildContentWhere(keyword, category, contentIDs)
	err := r.db.Where(cond, args...).
		Order("heat DESC, id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CountSearchNotes 统计命中笔记数
func (r *searchRepository) CountSearchNotes(keyword string, category int, contentIDs []uint) (int64, error) {
	var count int64
	cond, args := buildContentWhere(keyword, category, contentIDs)
	err := r.db.Model(&entity.Note{}).Where(cond, args...).Count(&count).Error
	return count, err
}

// SearchStrategies 搜索攻略（标题/分类/标签 OR 命中），按热度倒序，取 limit 条（limit=offset+size）
func (r *searchRepository) SearchStrategies(keyword string, category int, contentIDs []uint, limit int) ([]entity.Strategy, error) {
	var list []entity.Strategy
	cond, args := buildContentWhere(keyword, category, contentIDs)
	err := r.db.Where(cond, args...).
		Order("heat DESC, id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CountSearchStrategies 统计命中攻略数
func (r *searchRepository) CountSearchStrategies(keyword string, category int, contentIDs []uint) (int64, error) {
	var count int64
	cond, args := buildContentWhere(keyword, category, contentIDs)
	err := r.db.Model(&entity.Strategy{}).Where(cond, args...).Count(&count).Error
	return count, err
}

// SearchActivities 搜索活动（标题/分类/地点 OR 命中），按开始时间倒序，分页（与前台活动列表排序一致）
func (r *searchRepository) SearchActivities(keyword string, category int, offset, size int) ([]entity.Activity, error) {
	var list []entity.Activity
	cond, args := buildActivityWhere(keyword, category)
	err := r.db.Where(cond, args...).
		Order("start_time DESC, id DESC").
		Offset(offset).
		Limit(size).
		Find(&list).Error
	return list, err
}

// CountSearchActivities 统计命中活动数
func (r *searchRepository) CountSearchActivities(keyword string, category int) (int64, error) {
	var count int64
	cond, args := buildActivityWhere(keyword, category)
	err := r.db.Model(&entity.Activity{}).Where(cond, args...).Count(&count).Error
	return count, err
}

// buildContentWhere 构造笔记/攻略的搜索 WHERE 条件：status=1 AND (标题 LIKE OR 分类= OR id IN 内容ID)
// category>0 表示关键词命中分类枚举名，需精确匹配；contentIDs 非空表示关键词命中的标签所对应的内容ID（笔记/攻略ID）。
func buildContentWhere(keyword string, category int, contentIDs []uint) (string, []interface{}) {
	like := "%" + escapeLike(keyword) + "%"
	orConds := []string{"title LIKE ?"}
	orArgs := []interface{}{like}
	if category > 0 {
		orConds = append(orConds, "category = ?")
		orArgs = append(orArgs, category)
	}
	if len(contentIDs) > 0 {
		orConds = append(orConds, "id IN (?)")
		orArgs = append(orArgs, contentIDs)
	}
	cond := "status = ? AND (" + strings.Join(orConds, " OR ") + ")"
	args := append([]interface{}{1}, orArgs...)
	return cond, args
}

// buildActivityWhere 构造活动的搜索 WHERE 条件：is_del=0 AND (标题 LIKE OR 分类= OR 地点 LIKE)
func buildActivityWhere(keyword string, category int) (string, []interface{}) {
	like := "%" + escapeLike(keyword) + "%"
	orConds := []string{"title LIKE ?", "address LIKE ?"}
	orArgs := []interface{}{like, like}
	if category > 0 {
		orConds = append(orConds, "activity_type = ?")
		orArgs = append(orArgs, category)
	}
	cond := "is_del = ? AND (" + strings.Join(orConds, " OR ") + ")"
	args := append([]interface{}{0}, orArgs...)
	return cond, args
}

// escapeLike 转义 LIKE 通配符，避免用户输入 % _ 触发全量匹配（默认转义符为反斜杠）
func escapeLike(keyword string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(keyword)
}
