package tag

import (
	"context"

	response "tuyou/internal/model/dto/response/web/tag"
)

// TagService 标签服务接口
type TagService interface {
	// ListTags 获取启用标签列表
	ListTags(ctx context.Context) ([]response.TagItem, error)
}
