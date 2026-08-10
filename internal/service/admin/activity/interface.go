package activity

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/activity"
	response "tuyou/internal/model/dto/response/admin/activity"
)

// ActivityService 后台活动管理服务接口
type ActivityService interface {
	// ListActivities 分页查询活动列表
	ListActivities(ctx context.Context, req *request.ListActivityReq) (*response.ListActivityResp, error)
}
