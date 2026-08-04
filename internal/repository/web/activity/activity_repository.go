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

// FindByID 根据活动ID查询（未删除）
func (r *activityRepository) FindByID(activityID uint) (*entity.Activity, error) {
	var activity entity.Activity
	if err := r.db.Where("id = ? AND is_del = ?", activityID, 0).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// FindJoinedMembers 查询已报名成员列表（activity_join JOIN member，排除取消报名）
func (r *activityRepository) FindJoinedMembers(activityID uint) ([]entity.Member, error) {
	var members []entity.Member
	if err := r.db.Table("activity_join").
		Joins("JOIN member ON member.id = activity_join.member_id").
		Where("activity_join.activity_id = ?", activityID).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// ExistsJoin 检查用户是否已报名该活动
func (r *activityRepository) ExistsJoin(activityID uint, memberID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.ActivityJoin{}).
		Where("activity_id = ? AND member_id = ?", activityID, memberID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateActivityStatus 更新活动状态
func (r *activityRepository) UpdateActivityStatus(activityID uint, status int) error {
	return r.db.Model(&entity.Activity{}).Where("id = ?", activityID).
		UpdateColumn("activity_status", status).Error
}

// CreateJoin 创建报名记录
func (r *activityRepository) CreateJoin(tx *gorm.DB, activityID uint, memberID uint) error {
	return tx.Create(&entity.ActivityJoin{ActivityID: activityID, MemberID: memberID}).Error
}

// DeleteJoin 删除报名记录（硬删除，支持取消后重新报名）
func (r *activityRepository) DeleteJoin(tx *gorm.DB, activityID uint, memberID uint) error {
	return tx.Unscoped().Where("activity_id = ? AND member_id = ?", activityID, memberID).
		Delete(&entity.ActivityJoin{}).Error
}

// IncrCurrentPeople 报名人数 +1，返回实际影响行数（0=满员）
func (r *activityRepository) IncrCurrentPeople(tx *gorm.DB, activityID uint, maxPeople int) (int64, error) {
	result := tx.Model(&entity.Activity{}).
		Where("id = ? AND current_people < ?", activityID, maxPeople).
		UpdateColumn("current_people", gorm.Expr("current_people + 1"))
	return result.RowsAffected, result.Error
}

// DecrCurrentPeople 报名人数 -1
func (r *activityRepository) DecrCurrentPeople(tx *gorm.DB, activityID uint) error {
	return tx.Model(&entity.Activity{}).Where("id = ?", activityID).
		UpdateColumn("current_people", gorm.Expr("current_people - 1")).Error
}
