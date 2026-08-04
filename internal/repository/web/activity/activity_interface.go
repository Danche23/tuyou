package activity

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// ActivityRepository 活动仓储接口
type ActivityRepository interface {
	// Create 创建活动
	Create(tx *gorm.DB, activity *entity.Activity) error

	// GetAuthor 获取作者（领队）信息
	GetAuthor(memberID uint) (*entity.Member, error)

	// IncrPublishCount 发布计数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error
}
