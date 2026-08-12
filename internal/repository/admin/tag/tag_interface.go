package tag

import (
	"context"

	"tuyou/internal/model/entity"
)

// TagRepository 后台标签仓储接口
type TagRepository interface {
	// Count 统计标签数量（含筛选条件）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// FindList 分页查询标签列表（含筛选条件），返回实体
	FindList(ctx context.Context, filters map[string]interface{}, page, size int) ([]entity.Tag, error)
}
