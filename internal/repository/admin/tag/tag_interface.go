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

	// ExistsByName 标签名称是否已存在（Unscoped 包含禁用/软删标签）
	ExistsByName(ctx context.Context, tagName string) (bool, error)

	// Create 新增标签
	Create(ctx context.Context, tag *entity.Tag) error

	// FindByID 根据ID查询标签（Unscoped 包含禁用/软删标签），不存在返回 nil,nil
	FindByID(ctx context.Context, id uint) (*entity.Tag, error)

	// Update 更新标签（Unscoped 可更新禁用/软删标签）
	Update(ctx context.Context, id uint, updates map[string]interface{}) error

	// Delete 硬删除标签及其关联（单事务）
	Delete(ctx context.Context, id uint) error

	// BatchDelete 批量硬删除标签及其关联（单事务）；不存在的 ID 自动跳过
	BatchDelete(ctx context.Context, ids []uint) error
}
