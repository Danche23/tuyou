package activity

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 创建后台活动仓储实例
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

// applyFilters 拼装动态查询条件（针对 activity 表）
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["title"]; ok {
		query = query.Where("title LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["leader_name"]; ok {
		query = query.Joins("JOIN member ON member.id = activity.leader_member_id AND member.nickname LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["activity_type"]; ok {
		query = query.Where("activity_type = ?", v)
	}
	if v, ok := filters["difficulty"]; ok {
		query = query.Where("difficulty = ?", v)
	}
	if v, ok := filters["fee"]; ok {
		if v.(int) == 0 {
			query = query.Where("price = ?", 0)
		} else {
			query = query.Where("price > ?", 0)
		}
	}
	if v, ok := filters["activity_status"]; ok {
		query = query.Where("activity_status = ?", v)
	}
	if v, ok := filters["is_slide"]; ok {
		query = query.Where("is_slide = ?", v)
	}
	if v, ok := filters["begin_time"]; ok {
		query = query.Where("start_time >= ?", v)
	}
	if v, ok := filters["finish_time"]; ok {
		query = query.Where("start_time <= ?", v.(string)+" 23:59:59")
	}
	return query
}

// Count 统计活动数量
func (r *activityRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Activity{})
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindList 分页查询活动列表
func (r *activityRepository) FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Activity, error) {
	activities := make([]entity.Activity, 0)

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Activity{})
	query = applyFilters(query, filters)

	// 排序（sortField 来自 Service 白名单，安全）
	if sortField != "" {
		query = query.Order(sortField + " " + sortOrder)
	} else {
		query = query.Order("id DESC")
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

// FindByID 查询单个活动（Unscoped 含已删除/下架，不存在返回 nil, nil）
func (r *activityRepository) FindByID(ctx context.Context, activityID uint) (*entity.Activity, error) {
	var activity entity.Activity
	err := r.db.WithContext(ctx).Unscoped().First(&activity, activityID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// FindActivitiesByIDs 批量查询活动（Unscoped 含已删除/下架；IN 子句自动去重，不存在的 ID 自动跳过）
func (r *activityRepository) FindActivitiesByIDs(ctx context.Context, ids []uint) ([]entity.Activity, error) {
	activities := make([]entity.Activity, 0)
	if len(ids) == 0 {
		return activities, nil
	}
	if err := r.db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Order("id DESC").Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

// FindLeaderByName 按昵称查找领队（role=0 且 status=1 正常），不存在返回 (nil, nil)
func (r *activityRepository) FindLeaderByName(ctx context.Context, name string) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).
		Where("nickname = ? AND role = ? AND status = ?", name, 0, 1).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// FindLeaders 查询所有领队（role=0 且 status=1 正常），按 ID 升序
func (r *activityRepository) FindLeaders(ctx context.Context) ([]entity.Member, error) {
	leaders := make([]entity.Member, 0)
	if err := r.db.WithContext(ctx).
		Where("role = ?", 0).
		Where("status = ?", 1).
		Order("id ASC").
		Find(&leaders).Error; err != nil {
		return nil, err
	}
	return leaders, nil
}

// Update 更新活动字段（Unscoped 以支持恢复已下架/已删除活动）
func (r *activityRepository) Update(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Unscoped().Model(&entity.Activity{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 事务内硬删除活动及其关联数据（报名记录 + 主行）
func (r *activityRepository) Delete(tx *gorm.DB, activityID uint) error {
	// 关联数据：报名记录（活动无图片/标签/点赞/收藏表，评论 target_type 不含活动）
	if err := tx.Unscoped().Where("activity_id = ?", activityID).Delete(&entity.ActivityJoin{}).Error; err != nil {
		return err
	}
	// 主行
	return tx.Unscoped().Delete(&entity.Activity{}, activityID).Error
}
