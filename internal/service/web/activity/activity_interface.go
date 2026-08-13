package activity

import (
	"tuyou/internal/model/entity"

	requestActivity "tuyou/internal/model/dto/request/web/activity"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
)

// ActivityService 活动服务接口
type ActivityService interface {
	// CreateActivity 创建活动（发布活动，仅领队）
	CreateActivity(memberID uint, req *requestActivity.CreateActivityRequest) (*responseActivity.CreateActivityResponse, error)

	// GetActivityDetail 获取活动详情
	GetActivityDetail(activityID uint, memberID uint) (*responseActivity.ActivityDetailResponse, error)

	// ToggleSignup 报名/取消报名
	ToggleSignup(activityID uint, memberID uint) (*responseActivity.SignupToggleResponse, error)

	// DeleteActivity 删除活动（仅创建者）
	DeleteActivity(activityID uint, memberID uint) error

	// GetActivityList 获取活动列表（支持按分类筛选和分页）
	GetActivityList(memberID uint, req *requestActivity.ListActivityRequest) ([]responseActivity.ActivityListItem, int64, error)

	// BuildActivityItems 组装活动列表项（GetActivityList 与 search 模块共用的唯一实现）
	BuildActivityItems(memberID uint, activities []entity.Activity) ([]responseActivity.ActivityListItem, error)
}
