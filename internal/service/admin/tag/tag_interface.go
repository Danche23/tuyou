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

	// CreateTag 新增标签
	CreateTag(ctx context.Context, req *request.CreateTagReq) error

	// UpdateTag 编辑标签（状态/热门 必填；标签名称不可修改）
	UpdateTag(ctx context.Context, id uint, req *request.UpdateTagReq) error

	// DeleteTag 删除标签（硬删，连带清理关联）
	DeleteTag(ctx context.Context, id uint) error

	// BatchDeleteTags 批量删除标签（硬删，连带清理关联；不存在的 ID 自动跳过）
	BatchDeleteTags(ctx context.Context, req *request.BatchDeleteTagReq) error
}
