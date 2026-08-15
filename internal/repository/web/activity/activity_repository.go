package activity

import (
	"time"

	"tuyou/internal/model/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// DecrPublishCount 发布计数 -1
func (r *activityRepository) DecrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
}

// FindByID 根据活动ID查询（未删除）
func (r *activityRepository) FindByID(activityID uint) (*entity.Activity, error) {
	var activity entity.Activity
	if err := r.db.Where("id = ? AND is_del = ?", activityID, 0).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// FindJoinedMembers 查询已报名成员列表（activity_join JOIN member）
func (r *activityRepository) FindJoinedMembers(activityID uint) ([]entity.Member, error) {
	var members []entity.Member
	if err := r.db.Table("activity_join").
		Select("member.id, member.nickname, member.avatar").
		Joins("JOIN member ON member.id = activity_join.member_id").
		Where("activity_join.activity_id = ? AND member.delete_time IS NULL", activityID).
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

// ExistsFollow 检查当前用户是否已关注领队
func (r *activityRepository) ExistsFollow(followerID uint, followeeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
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

// CreateJoin 创建报名记录（INSERT IGNORE：重复报名不报错，返回是否新插入）
// 并发双击时唯一索引兜底：第二次插入被忽略，返回 false，事务外层不再加人数、不发通知
func (r *activityRepository) CreateJoin(tx *gorm.DB, activityID uint, memberID uint) (bool, error) {
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entity.ActivityJoin{ActivityID: activityID, MemberID: memberID})
	return result.RowsAffected > 0, result.Error
}

// DeleteJoin 删除报名记录（硬删除，支持取消后重新报名）；返回是否实际删除了记录
func (r *activityRepository) DeleteJoin(tx *gorm.DB, activityID uint, memberID uint) (bool, error) {
	res := tx.Unscoped().Where("activity_id = ? AND member_id = ?", activityID, memberID).
		Delete(&entity.ActivityJoin{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
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

// GetCurrentPeople 查询活动当前报名人数（返回 DB 真实值，供报名/取消后响应使用）
func (r *activityRepository) GetCurrentPeople(db *gorm.DB, activityID uint) (int, error) {
	var activity entity.Activity
	if err := db.Select("current_people").First(&activity, activityID).Error; err != nil {
		return 0, err
	}
	return activity.CurrentPeople, nil
}

// SoftDelete 逻辑删除活动（is_del=1 + delete_time，与笔记/攻略软删语义对齐）
func (r *activityRepository) SoftDelete(tx *gorm.DB, activityID uint) error {
	return tx.Unscoped().Model(&entity.Activity{}).Where("id = ?", activityID).
		Updates(map[string]interface{}{
			"is_del":      1,
			"delete_time": time.Now(),
		}).Error
}

// FindList 查询活动列表（分页，按分类筛选，过滤已删除，按开始时间倒序）
func (r *activityRepository) FindList(category int, page, size int) ([]entity.Activity, int64, error) {
	var list []entity.Activity
	var total int64

	db := r.db.Model(&entity.Activity{}).Where("is_del = ?", 0)
	if category > 0 {
		db = db.Where("activity_type = ?", category)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Offset(offset).Limit(size).Order("start_time DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetMembersByIDs 批量获取成员信息
func (r *activityRepository) GetMembersByIDs(ids []uint) ([]entity.Member, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var members []entity.Member
	if err := r.db.Where("id IN ?", ids).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// FindJoinedActivityIDs 查询当前用户已报名的活动ID集合
func (r *activityRepository) FindJoinedActivityIDs(activityIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if len(activityIDs) == 0 {
		return result, nil
	}

	var joins []entity.ActivityJoin
	if err := r.db.Model(&entity.ActivityJoin{}).
		Where("activity_id IN ? AND member_id = ?", activityIDs, memberID).
		Find(&joins).Error; err != nil {
		return nil, err
	}

	for _, j := range joins {
		result[j.ActivityID] = true
	}
	return result, nil
}
