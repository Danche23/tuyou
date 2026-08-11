package tag

import (
	"context"

	"tuyou/internal/model/entity"
)

// TagRepository 标签仓储接口
type TagRepository interface {
	// ListEnabledTags 查询全部启用标签（status=1），按 id 升序
	ListEnabledTags(ctx context.Context) ([]entity.Tag, error)
}
