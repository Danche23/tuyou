package tag

import (
	"context"
	"errors"

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

// ExistsByName 标签名称是否已存在（Unscoped 包含禁用/软删标签，名字仍被占用）
func (r *tagRepository) ExistsByName(ctx context.Context, tagName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&entity.Tag{}).
		Where("tag_name = ?", tagName).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create 新增标签
func (r *tagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// FindByID 根据ID查询标签（Unscoped 包含禁用/软删标签），不存在返回 nil,nil
func (r *tagRepository) FindByID(ctx context.Context, id uint) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).Unscoped().Model(&entity.Tag{}).
		Where("id = ?", id).
		First(&tag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// Update 更新标签（Unscoped 可更新禁用/软删标签）
func (r *tagRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Unscoped().Model(&entity.Tag{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// deleteTagInTx 事务内硬删单个标签及其关联（关联表清理 + 主行硬删）
func deleteTagInTx(tx *gorm.DB, id uint) error {
	// 关联表：笔记/攻略引用该标签的记录一并清理
	if err := tx.Unscoped().Where("tag_id = ?", id).Delete(&entity.NoteTag{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("tag_id = ?", id).Delete(&entity.StrategyTag{}).Error; err != nil {
		return err
	}
	// 主行硬删（Unscoped 绕过软删，直接移除行）
	return tx.Unscoped().Delete(&entity.Tag{}, id).Error
}

// Delete 硬删除标签及其关联（单事务）
func (r *tagRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteTagInTx(tx, id)
	})
}

// BatchDelete 批量硬删除标签及其关联（单事务）；不存在的 ID 删除 0 行自动跳过
func (r *tagRepository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			if err := deleteTagInTx(tx, id); err != nil {
				return err
			}
		}
		return nil
	})
}
