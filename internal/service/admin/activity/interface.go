package activity

import (
	"context"

	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/activity"
	response "tuyou/internal/model/dto/response/admin/activity"
	"tuyou/internal/model/entity"
)

// webActivityWriter 复用前台活动仓库的通用写操作（同一张表的操作不重复实现）
type webActivityWriter interface {
	// Create 创建活动
	Create(tx *gorm.DB, activity *entity.Activity) error
	// GetMembersByIDs 批量获取成员信息
	GetMembersByIDs(ids []uint) ([]entity.Member, error)
}

// ActivityService 后台活动管理服务接口
type ActivityService interface {
	// ListActivities 分页查询活动列表
	ListActivities(ctx context.Context, req *request.ListActivityReq) (*response.ListActivityResp, error)

	// CreateActivity 创建活动
	CreateActivity(ctx context.Context, req *request.CreateActivityReq) error

	// GetActivityDetail 获取活动详情
	GetActivityDetail(ctx context.Context, activityID uint) (*response.ActivityDetailResp, error)

	// UpdateActivity 编辑活动（分类/轮播/状态三字段整体覆盖，返回更新后的详情）
	UpdateActivity(ctx context.Context, activityID uint, req *request.UpdateActivityReq) (*response.ActivityDetailResp, error)

	// DeleteActivity 删除活动（硬删，连带报名记录）
	DeleteActivity(ctx context.Context, activityID uint) error

	// BatchDeleteActivities 批量硬删除活动（连带报名记录；不存在的 ID 自动跳过）
	BatchDeleteActivities(ctx context.Context, req *request.BatchDeleteReq) error

	// ExportActivities 按 ID 批量导出活动列表项（不含软删过滤，已下架也导出）
	ExportActivities(ctx context.Context, req *request.ExportActivityReq) ([]response.ActivityListItem, error)

	// ListLeaderNames 获取领队名称列表
	ListLeaderNames(ctx context.Context) ([]response.LeaderNameItem, error)
}
