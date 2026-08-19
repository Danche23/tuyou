package activity

import (
	"time"

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

	// ExistsFollow 检查用户是否已关注领队
	ExistsFollow(followerID uint, followeeID uint) (bool, error)

	// UpdateActivityStatus 更新活动状态
	UpdateActivityStatus(activityID uint, status int) error

	// CreateJoin 创建报名记录（INSERT IGNORE 兜底：重复报名返回 false 且不报错）
	CreateJoin(tx *gorm.DB, activityID uint, memberID uint) (bool, error)

	// DeleteJoin 删除报名记录（硬删除，支持取消后重新报名）；返回是否实际删除了记录
	DeleteJoin(tx *gorm.DB, activityID uint, memberID uint) (bool, error)

	// IncrCurrentPeople 报名人数 +1，返回实际影响行数（0=满员）
	IncrCurrentPeople(tx *gorm.DB, activityID uint, maxPeople int) (int64, error)

	// DecrCurrentPeople 报名人数 -1
	DecrCurrentPeople(tx *gorm.DB, activityID uint) error

	// GetCurrentPeople 查询活动当前报名人数（返回 DB 真实值）
	GetCurrentPeople(db *gorm.DB, activityID uint) (int, error)

	// SoftDelete 逻辑删除活动（is_del=1）
	SoftDelete(tx *gorm.DB, activityID uint) error

	// FindList 查询活动列表（分页，按分类/日期筛选，过滤已删除）
	FindList(category int, page, size int, dateFilter time.Time) ([]entity.Activity, int64, error)

	// GetMembersByIDs 批量获取成员信息
	GetMembersByIDs(ids []uint) ([]entity.Member, error)

	// FindJoinedActivityIDs 查询当前用户已报名的活动ID集合
	FindJoinedActivityIDs(activityIDs []uint, memberID uint) (map[uint]bool, error)
}
