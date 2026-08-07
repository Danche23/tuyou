package guide

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/guide"
	response "tuyou/internal/model/dto/response/admin/guide"
)

// GuideService 后台攻略管理服务接口
type GuideService interface {
	// ListGuides 分页查询攻略列表
	ListGuides(ctx context.Context, req *request.ListGuideReq) (*response.ListGuideResp, error)

	// GetGuideDetail 获取攻略详情
	GetGuideDetail(ctx context.Context, id uint) (*response.GuideDetailResp, error)

	// UpdateGuide 编辑攻略（分类/轮播/状态三字段必传，整体覆盖），返回更新后的详情
	UpdateGuide(ctx context.Context, id uint, req *request.UpdateGuideReq) (*response.GuideDetailResp, error)
}
