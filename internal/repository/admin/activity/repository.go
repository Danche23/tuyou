package activity

import (
	"context"

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

// FindMembersByIDs 批量查询领队信息
func (r *activityRepository) FindMembersByIDs(ctx context.Context, memberIDs []uint) ([]entity.Member, error) {
	if len(memberIDs) == 0 {
		return []entity.Member{}, nil
	}

	members := make([]entity.Member, 0)
	if err := r.db.WithContext(ctx).Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
