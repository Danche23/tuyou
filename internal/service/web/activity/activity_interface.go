package activity

import (
	requestActivity "tuyou/internal/model/dto/request/web/activity"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
)

// ActivityService 活动服务接口
type ActivityService interface {
	// CreateActivity 创建活动（发布活动，仅领队）
	CreateActivity(memberID uint, req *requestActivity.CreateActivityRequest) (*responseActivity.CreateActivityResponse, error)

	// GetActivityDetail 获取活动详情
	GetActivityDetail(activityID uint, memberID uint) (*responseActivity.ActivityDetailResponse, error)
}
