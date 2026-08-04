package guide

import (
	requestGuide "tuyou/internal/model/dto/request/web/guide"
	responseGuide "tuyou/internal/model/dto/response/web/guide"
)

// GuideService 攻略服务接口（当前仅覆盖发布流程，后续接口按需扩展）
type GuideService interface {
	// CreateGuide 发布攻略
	CreateGuide(memberID uint, req *requestGuide.CreateGuideRequest) (*responseGuide.CreateGuideResponse, error)

	// GetGuideDetail 获取攻略详情
	GetGuideDetail(strategyID uint, memberID uint) (*responseGuide.GuideDetailResponse, error)

	// ToggleLikeGuide 点赞/取消点赞
	ToggleLikeGuide(strategyID uint, memberID uint) (*responseGuide.LikeToggleResponse, error)

	// ToggleCollectGuide 收藏/取消收藏
	ToggleCollectGuide(strategyID uint, memberID uint) (*responseGuide.CollectToggleResponse, error)

	// DeleteGuide 删除攻略（下架，仅作者）
	DeleteGuide(strategyID uint, memberID uint) error
}
