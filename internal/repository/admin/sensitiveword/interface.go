package sensitiveword

import (
	"context"

	"tuyou/internal/model/entity"
)

// SensitiveWordRepository 后台敏感词仓储接口
type SensitiveWordRepository interface {
	// Count 统计敏感词数量（含筛选条件；Unscoped 硬删语义下行已物理删，带上无副作用）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// FindList 分页查询敏感词列表（含筛选条件；Unscoped），按 id 升序
	FindList(ctx context.Context, filters map[string]interface{}, page, size int) ([]entity.SensitiveWord, error)

	// ExistsByName 敏感词是否已存在（Unscoped；词唯一索引，重名拦截）
	ExistsByName(ctx context.Context, word string) (bool, error)

	// Create 新增敏感词
	Create(ctx context.Context, word *entity.SensitiveWord) error

	// FindByID 根据ID查询敏感词（Unscoped），不存在返回 nil,nil
	FindByID(ctx context.Context, id uint) (*entity.SensitiveWord, error)

	// Delete 硬删除敏感词（Unscoped，物理删行）
	Delete(ctx context.Context, id uint) error

	// ListActiveWords 取启用词（status=1）供 trie 热更
	ListActiveWords(ctx context.Context) ([]string, error)
}
