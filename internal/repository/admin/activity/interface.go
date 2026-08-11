package activity

import (
	"context"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

// ActivityRepository 后台活动管理仓储接口
type ActivityRepository interface {
	// Count 统计活动数量（Unscoped 含已删除）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// FindList 分页查询活动列表（Unscoped 含已删除）
	FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Activity, error)

	// FindByID 查询单个活动（Unscoped 含已删除/下架，不存在返回 nil, nil）
	FindByID(ctx context.Context, activityID uint) (*entity.Activity, error)

	// FindActivitiesByIDs 批量查询活动（Unscoped 含已删除/下架；IN 子句自动去重，不存在的 ID 自动跳过）
	FindActivitiesByIDs(ctx context.Context, ids []uint) ([]entity.Activity, error)

	// FindLeaderByName 按昵称查找领队（role=0 且正常），不存在返回 (nil, nil)
	FindLeaderByName(ctx context.Context, name string) (*entity.Member, error)

	// FindLeaders 查询所有领队（role=0 且正常状态），用于活动列表按领队名称筛选
	FindLeaders(ctx context.Context) ([]entity.Member, error)

	// Update 更新活动字段（Unscoped 以支持恢复已下架/已删除活动）
	Update(tx *gorm.DB, id uint, updates map[string]interface{}) error

	// Delete 事务内硬删除活动及其关联数据（报名记录 + 主行）
	Delete(tx *gorm.DB, activityID uint) error
}
