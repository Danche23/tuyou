package activity

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 创建活动仓储实例
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

// Create 创建活动
func (r *activityRepository) Create(tx *gorm.DB, activity *entity.Activity) error {
	return tx.Create(activity).Error
}

// GetAuthor 获取作者（领队）信息
func (r *activityRepository) GetAuthor(memberID uint) (*entity.Member, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// IncrPublishCount 发布计数 +1
func (r *activityRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}
