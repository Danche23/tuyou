package tag

import (
	"context"

	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓储实例
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// ListEnabledTags 查询全部启用标签
func (r *tagRepository) ListEnabledTags(ctx context.Context) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := r.db.WithContext(ctx).Where("status = ?", 1).Order("id ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
