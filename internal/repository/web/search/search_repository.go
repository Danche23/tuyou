package search

import (
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
