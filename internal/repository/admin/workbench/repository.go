package workbench

import (
	"context"
	"time"

	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

type workbenchRepository struct {
	db *gorm.DB
}

func NewWorkbenchRepository(db *gorm.DB) WorkbenchRepository {
	return &workbenchRepository{db: db}
}

// CountMembers 用户总数（GORM 自动过滤软删除）
func (r *workbenchRepository) CountMembers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Member{}).Count(&count).Error
	return count, err
}

// CountActivities 活动总数（activity 用自身逻辑删除字段 is_del）
func (r *workbenchRepository) CountActivities(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Activity{}).Where("is_del = ?", 0).Count(&count).Error
	return count, err
}

// CountNotes 笔记总数
func (r *workbenchRepository) CountNotes(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Note{}).Count(&count).Error
	return count, err
}

// CountStrategies 攻略总数
func (r *workbenchRepository) CountStrategies(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Strategy{}).Count(&count).Error
	return count, err
}

// CountComments 评论总数（仅正常 status=1）
func (r *workbenchRepository) CountComments(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Comment{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

// CountRegistersByDay 某天注册人数，区间 [day, day+1)
func (r *workbenchRepository) CountRegistersByDay(ctx context.Context, day time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Member{}).
		Where("create_time >= ? AND create_time < ?", day, day.AddDate(0, 0, 1)).
		Count(&count).Error
	return count, err
}
