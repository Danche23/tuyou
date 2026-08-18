package sensitiveword

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

type sensitiveWordRepository struct {
	db *gorm.DB
}

func NewSensitiveWordRepository(db *gorm.DB) SensitiveWordRepository {
	return &sensitiveWordRepository{db: db}
}

func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["keyword"]; ok {
		query = query.Where("word LIKE ?", "%"+v.(string)+"%")
	}
	return query
}

func (r *sensitiveWordRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.SensitiveWord{})
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *sensitiveWordRepository) FindList(ctx context.Context, filters map[string]interface{}, page, size int) ([]entity.SensitiveWord, error) {
	words := make([]entity.SensitiveWord, 0)
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.SensitiveWord{})
	query = applyFilters(query, filters)
	offset := (page - 1) * size
	if err := query.Order("id ASC").Offset(offset).Limit(size).Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (r *sensitiveWordRepository) ExistsByName(ctx context.Context, word string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&entity.SensitiveWord{}).
		Where("word = ?", word).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *sensitiveWordRepository) Create(ctx context.Context, word *entity.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *sensitiveWordRepository) FindByID(ctx context.Context, id uint) (*entity.SensitiveWord, error) {
	var w entity.SensitiveWord
	err := r.db.WithContext(ctx).Unscoped().Model(&entity.SensitiveWord{}).
		Where("id = ?", id).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *sensitiveWordRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.SensitiveWord{}, id).Error
}

// ListActiveWords 取启用词（status=1）供 trie 热更
func (r *sensitiveWordRepository) ListActiveWords(ctx context.Context) ([]string, error) {
	var words []string
	err := r.db.WithContext(ctx).Model(&entity.SensitiveWord{}).
		Where("status = ?", 1).Pluck("word", &words).Error
	return words, err
}
