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

// SearchNotes 搜索笔记（标题/正文/分类 OR 命中），相关度优先、热度兜底，取 limit 条（limit=offset+size）
func (r *searchRepository) SearchNotes(keyword string, category int, limit int) ([]entity.Note, error) {
	var list []entity.Note
	cond, args := buildContentWhere(keyword, category)
	err := r.db.Where(cond, args...).
		Order(gorm.Expr("MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE) DESC", keyword)).
		Order("heat DESC, id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CountSearchNotes 统计命中笔记数
func (r *searchRepository) CountSearchNotes(keyword string, category int) (int64, error) {
	var count int64
	cond, args := buildContentWhere(keyword, category)
	err := r.db.Model(&entity.Note{}).Where(cond, args...).Count(&count).Error
	return count, err
}

// SearchStrategies 搜索攻略（标题/正文/分类 OR 命中），相关度优先、热度兜底，取 limit 条（limit=offset+size）
func (r *searchRepository) SearchStrategies(keyword string, category int, limit int) ([]entity.Strategy, error) {
	var list []entity.Strategy
	cond, args := buildContentWhere(keyword, category)
	err := r.db.Where(cond, args...).
		Order(gorm.Expr("MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE) DESC", keyword)).
		Order("heat DESC, id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// CountSearchStrategies 统计命中攻略数
func (r *searchRepository) CountSearchStrategies(keyword string, category int) (int64, error) {
	var count int64
	cond, args := buildContentWhere(keyword, category)
	err := r.db.Model(&entity.Strategy{}).Where(cond, args...).Count(&count).Error
	return count, err
}

// SearchActivities 搜索活动（标题/地点/分类 OR 命中），相关度优先、开始时间兜底，分页
func (r *searchRepository) SearchActivities(keyword string, category int, offset, size int) ([]entity.Activity, error) {
	var list []entity.Activity
	cond, args := buildActivityWhere(keyword, category)
	err := r.db.Where(cond, args...).
		Order(gorm.Expr("MATCH(title, address) AGAINST(? IN NATURAL LANGUAGE MODE) DESC", keyword)).
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

// buildContentWhere 构造笔记/攻略的搜索 WHERE 条件：status=1 AND (标题/正文全文命中 OR 分类=)
// 关键词匹配走 MATCH...AGAINST 全文检索（FULLTEXT+ngram 倒排索引，带相关度分）；
// category>0 表示关键词命中分类枚举名，需精确匹配。搜索词长度已在 service 层拦截（≥2 字）。
func buildContentWhere(keyword string, category int) (string, []interface{}) {
	var orConds []string
	var orArgs []interface{}
	orConds = append(orConds, "MATCH(title, content) AGAINST(? IN NATURAL LANGUAGE MODE)")
	orArgs = append(orArgs, keyword)
	if category > 0 {
		orConds = append(orConds, "category = ?")
		orArgs = append(orArgs, category)
	}
	cond := "status = ? AND (" + strings.Join(orConds, " OR ") + ")"
	args := append([]interface{}{1}, orArgs...)
	return cond, args
}

// buildActivityWhere 构造活动的搜索 WHERE 条件：is_del=0 AND (标题/地点全文命中 OR 分类=)
// 关键词匹配走 MATCH...AGAINST（标题+地点字段）；搜索词长度已在 service 层拦截（≥2 字）。
func buildActivityWhere(keyword string, category int) (string, []interface{}) {
	var orConds []string
	var orArgs []interface{}
	orConds = append(orConds, "MATCH(title, address) AGAINST(? IN NATURAL LANGUAGE MODE)")
	orArgs = append(orArgs, keyword)
	if category > 0 {
		orConds = append(orConds, "activity_type = ?")
		orArgs = append(orArgs, category)
	}
	cond := "is_del = ? AND (" + strings.Join(orConds, " OR ") + ")"
	args := append([]interface{}{0}, orArgs...)
	return cond, args
}
