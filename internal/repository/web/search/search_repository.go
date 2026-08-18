package search

import (
	"strings"
	"time"

	"tuyou/internal/model/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// HistoryLimit 每个用户最多保留的搜索历史条数（导出，供 service 层查询历史复用同一上限）
const HistoryLimit = 20

// UpsertSearchHistory 记录一条搜索历史：同词去重置顶（刷新 update_time），超出 HistoryLimit 条淘汰最旧
// create_time 保留首次搜索时间不变；update_time 作为「最近搜索时间」用于置顶排序。
func (r *searchRepository) UpsertSearchHistory(memberID uint, keyword string) error {
	history := &entity.SearchHistory{MemberID: memberID, Keyword: keyword}
	// 唯一键 (member_id, keyword) 命中时只刷新 update_time，实现「同词置顶」
	// 注意：不能用 gorm.Expr("NOW()")——本 MySQL 服务器时区为 UTC，NOW() 会写成 8 小时前的
	// UTC 时间，与 GORM 自动时间戳（Go 本地时间）不一致，导致置顶看起来失效。用 time.Now() 同源。
	if err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "member_id"}, {Name: "keyword"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"update_time": time.Now(),
		}),
	}).Create(history).Error; err != nil {
		return err
	}
	// 截断：只保留最近 HistoryLimit 条（第 21 条起的全部删掉）。
	// 用「多表 DELETE + JOIN 子查询」而非 WHERE id IN (同表子查询)：
	// 后者在 MySQL 报错误 1093「不能修改子查询里选出的表」，JOIN 形式是官方规避写法。
	return r.db.Exec(
		"DELETE sh FROM search_history sh "+
			"INNER JOIN (SELECT id FROM search_history WHERE member_id = ? ORDER BY update_time DESC, id DESC LIMIT ? OFFSET ?) old ON old.id = sh.id",
		memberID, HistoryLimit, HistoryLimit,
	).Error
}

// ListSearchHistory 查询某用户最近搜索历史（update_time 倒序，最多 limit 条），返回含 id 的完整记录
func (r *searchRepository) ListSearchHistory(memberID uint, limit int) ([]entity.SearchHistory, error) {
	var list []entity.SearchHistory
	err := r.db.Where("member_id = ?", memberID).
		Order("update_time DESC, id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// DeleteSearchHistoryByID 删除单条搜索历史（硬删；限定 member_id，只能删自己的；幂等）
// 用硬删而非软删：软删的行仍占着唯一索引 (member_id, keyword)，同词再搜会插不进。
func (r *searchRepository) DeleteSearchHistoryByID(memberID, id uint) error {
	return r.db.Unscoped().
		Where("id = ? AND member_id = ?", id, memberID).
		Delete(&entity.SearchHistory{}).Error
}

// ClearSearchHistory 清空某用户全部搜索历史（硬删；幂等）
func (r *searchRepository) ClearSearchHistory(memberID uint) error {
	return r.db.Unscoped().
		Where("member_id = ?", memberID).
		Delete(&entity.SearchHistory{}).Error
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
