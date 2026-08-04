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

	// DecrPublishCount 发布计数 -1
	DecrPublishCount(tx *gorm.DB, memberID uint) error

	// FindByID 根据活动ID查询（未删除）
	FindByID(activityID uint) (*entity.Activity, error)

	// FindJoinedMembers 查询已报名成员列表（activity_join JOIN member）
	FindJoinedMembers(activityID uint) ([]entity.Member, error)

	// ExistsJoin 检查用户是否已报名该活动
	ExistsJoin(activityID uint, memberID uint) (bool, error)

	// UpdateActivityStatus 更新活动状态
	UpdateActivityStatus(activityID uint, status int) error

	// CreateJoin 创建报名记录
	CreateJoin(tx *gorm.DB, activityID uint, memberID uint) error

	// DeleteJoin 删除报名记录（硬删除，支持取消后重新报名）
	DeleteJoin(tx *gorm.DB, activityID uint, memberID uint) error

	// IncrCurrentPeople 报名人数 +1，返回实际影响行数（0=满员）
	IncrCurrentPeople(tx *gorm.DB, activityID uint, maxPeople int) (int64, error)

	// DecrCurrentPeople 报名人数 -1
	DecrCurrentPeople(tx *gorm.DB, activityID uint) error

	// SoftDelete 逻辑删除活动（is_del=1）
	SoftDelete(tx *gorm.DB, activityID uint) error
}
