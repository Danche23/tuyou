package guide

import (
	"context"

	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// GuideRepository 攻略仓储接口（当前仅覆盖发布流程，后续接口按需扩展）
type GuideRepository interface {
	// Create 创建攻略
	Create(tx *gorm.DB, strategy *entity.Strategy) error

	// FindTags 根据标签名查找
	FindTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error)

	// CreateStrategyTags 批量创建攻略-标签关联
	CreateStrategyTags(tx *gorm.DB, strategyID uint, tagIDs []uint) error

	// CreateStrategyImages 批量创建攻略图片
	CreateStrategyImages(tx *gorm.DB, strategyID uint, imageURLs []string) error

	// IncrPublishCount 发布计数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error

	// FindByID 根据攻略ID查攻略主表（仅上架）
	FindByID(strategyID uint) (*entity.Strategy, error)

	// FindImagesByStrategyID 查攻略图片列表
	FindImagesByStrategyID(strategyID uint) ([]entity.StrategyImage, error)

	// FindTagsByStrategyID 查攻略标签列表
	FindTagsByStrategyID(strategyID uint) ([]entity.Tag, error)

	// GetAuthor 获取作者信息
	GetAuthor(memberID uint) (*entity.Member, error)

	// ExistsFollow 检查是否已关注作者
	ExistsFollow(followerID uint, followeeID uint) (bool, error)

	// IncrViewCount 浏览量 +1（降级直写 DB）
	IncrViewCount(strategyID uint) error

	// XAddViewCount 发送浏览量消息到 Redis Stream（异步批量写 DB）
	XAddViewCount(ctx context.Context, strategyID uint) error

	// ExistsLike 检查用户是否已点赞该攻略
	ExistsLike(strategyID uint, memberID uint) (bool, error)

	// CreateLike 创建点赞记录
	CreateLike(tx *gorm.DB, strategyID uint, memberID uint) error

	// DeleteLike 删除点赞记录（硬删除）
	DeleteLike(tx *gorm.DB, strategyID uint, memberID uint) error

	// IncrLikeCount 攻略点赞数 +1
	IncrLikeCount(tx *gorm.DB, strategyID uint) error

	// DecrLikeCount 攻略点赞数 -1
	DecrLikeCount(tx *gorm.DB, strategyID uint) error

	// GetCachedLikeStatus 读取点赞状态缓存（found=false 表示缓存未命中）
	GetCachedLikeStatus(ctx context.Context, strategyID, memberID uint) (isLiked bool, found bool, err error)

	// SetCachedLikeStatus 写入点赞状态缓存
	SetCachedLikeStatus(ctx context.Context, strategyID, memberID uint, isLiked bool)

	// GetCachedLikeCount 读取点赞数缓存
	GetCachedLikeCount(ctx context.Context, strategyID uint) (int64, error)

	// SetCachedLikeCount 写入点赞数缓存
	SetCachedLikeCount(ctx context.Context, strategyID uint, count int64)

	// ExistsCollect 检查用户是否已收藏该攻略
	ExistsCollect(strategyID uint, memberID uint) (bool, error)

	// CreateCollect 创建收藏记录
	CreateCollect(tx *gorm.DB, strategyID uint, memberID uint) error

	// DeleteCollect 删除收藏记录（硬删除）
	DeleteCollect(tx *gorm.DB, strategyID uint, memberID uint) error

	// IncrCollectCount 攻略收藏数 +1
	IncrCollectCount(tx *gorm.DB, strategyID uint) error

	// DecrCollectCount 攻略收藏数 -1
	DecrCollectCount(tx *gorm.DB, strategyID uint) error

	// GetCachedCollectStatus 读取收藏状态缓存（found=false 表示缓存未命中）
	GetCachedCollectStatus(ctx context.Context, strategyID, memberID uint) (isFavorited bool, found bool, err error)

	// SetCachedCollectStatus 写入收藏状态缓存
	SetCachedCollectStatus(ctx context.Context, strategyID, memberID uint, isFavorited bool)
}
