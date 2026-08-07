package guide

import (
	"context"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

// GuideRepository 后台攻略管理仓储接口
type GuideRepository interface {
	// Count 统计攻略数量（Unscoped 含软删）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// FindList 分页查询攻略列表（Unscoped 含软删）
	FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Strategy, error)

	// FindByID 根据 ID 查询单篇攻略（Unscoped 含软删，不存在返回 nil, nil）
	FindByID(ctx context.Context, id uint) (*entity.Strategy, error)

	// FindTagsByStrategyIDs 批量查询攻略标签 map[strategyID][]tagName
	FindTagsByStrategyIDs(ctx context.Context, strategyIDs []uint) (map[uint][]string, error)

	// FindMembersByIDs 批量查询作者信息
	FindMembersByIDs(ctx context.Context, memberIDs []uint) ([]entity.Member, error)

	// FindCoverImagesByStrategyIDs 批量查询攻略封面图（每篇第一张）map[strategyID]url
	FindCoverImagesByStrategyIDs(ctx context.Context, strategyIDs []uint) (map[uint]string, error)

	// Update 更新攻略字段（updates map 中的 key 都会被更新；Unscoped 以支持恢复已软删攻略）
	Update(tx *gorm.DB, id uint, updates map[string]interface{}) error

	// UpdateCommentStatusByTarget 更新目标攻略/笔记下的评论状态
	UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error

	// IncrPublishCount 作者发布数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error

	// DecrPublishCount 作者发布数 -1
	DecrPublishCount(tx *gorm.DB, memberID uint) error
}
