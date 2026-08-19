package sensitiveword

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/sensitiveword"
	response "tuyou/internal/model/dto/response/admin/sensitiveword"
)

// SensitiveWordService 后台敏感词管理服务接口
type SensitiveWordService interface {
	List(ctx context.Context, req *request.ListReq) (*response.ListResp, error)
	Create(ctx context.Context, req *request.CreateReq) error
	Delete(ctx context.Context, id uint) error
}
