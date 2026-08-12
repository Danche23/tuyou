package tag

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/tag"
	response "tuyou/internal/model/dto/response/admin/tag"
)

// TagService 后台标签管理服务接口
type TagService interface {
	// ListTags 分页查询标签列表
	ListTags(ctx context.Context, req *request.ListTagReq) (*response.ListTagResp, error)
}
