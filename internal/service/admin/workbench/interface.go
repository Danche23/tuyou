package workbench

import (
	"context"

	response "tuyou/internal/model/dto/response/admin/workbench"
)

// WorkbenchService 后台工作台服务
type WorkbenchService interface {
	// GetWorkbench 获取工作台统计数据
	GetWorkbench(ctx context.Context) (*response.WorkbenchResp, error)
}
