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

	// FindByID 根据活动ID查询（未删除）
	FindByID(activityID uint) (*entity.Activity, error)

	// FindJoinedMembers 查询已报名成员列表（activity_join JOIN member）
	FindJoinedMembers(activityID uint) ([]entity.Member, error)

	// ExistsJoin 检查用户是否已报名该活动
	ExistsJoin(activityID uint, memberID uint) (bool, error)

	// UpdateActivityStatus 更新活动状态
	UpdateActivityStatus(activityID uint, status int) error
}
