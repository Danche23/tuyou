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

// CountMembers 用户总数（含软删除 delete_time 置位，硬删除不在表中）
func (r *workbenchRepository) CountMembers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Member{}).Unscoped().Count(&count).Error
	return count, err
}

// CountActivities 活动总数（含逻辑删除 is_del=1 与软删 delete_time，硬删除不在表中）
func (r *workbenchRepository) CountActivities(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Activity{}).Unscoped().Count(&count).Error
	return count, err
}

// CountNotes 笔记总数（含软删除 delete_time，硬删除不在表中）
func (r *workbenchRepository) CountNotes(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Note{}).Unscoped().Count(&count).Error
	return count, err
}

// CountStrategies 攻略总数（含软删除 delete_time，硬删除不在表中）
func (r *workbenchRepository) CountStrategies(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Strategy{}).Unscoped().Count(&count).Error
	return count, err
}

// CountComments 评论总数（含下架 status=0 与软删 delete_time，硬删除不在表中）
func (r *workbenchRepository) CountComments(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Comment{}).Unscoped().Count(&count).Error
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
