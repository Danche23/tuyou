package tag

import (
	"context"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建后台标签仓储实例
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// applyFilters 拼装动态查询条件（针对 tag 表）
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["tag_name"]; ok {
		query = query.Where("tag_name LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["status"]; ok {
		query = query.Where("status = ?", v)
	}
	if v, ok := filters["begin_time"]; ok {
		query = query.Where("create_time >= ?", v)
	}
	if v, ok := filters["finish_time"]; ok {
		query = query.Where("create_time <= ?", v)
	}
	return query
}

// Count 统计标签数量（含筛选条件；Unscoped 包含禁用/软删标签）
func (r *tagRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Tag{})
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindList 分页查询标签列表（含筛选条件；Unscoped 包含禁用/软删标签），按 id 升序
func (r *tagRepository) FindList(ctx context.Context, filters map[string]interface{}, page, size int) ([]entity.Tag, error) {
	tags := make([]entity.Tag, 0)

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Tag{})
	query = applyFilters(query, filters)
	query = query.Order("id ASC")

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
