package activity

import (
	"context"

	"tuyou/internal/model/entity"
)

// ActivityRepository 后台活动管理仓储接口
type ActivityRepository interface {
	// Count 统计活动数量（Unscoped 含已删除）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// FindList 分页查询活动列表（Unscoped 含已删除）
	FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Activity, error)

	// FindMembersByIDs 批量查询领队信息
	FindMembersByIDs(ctx context.Context, memberIDs []uint) ([]entity.Member, error)
}
