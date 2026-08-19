package guide

import (
	"context"
	"fmt"
	"time"

	"tuyou/internal/model/entity"
	"tuyou/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type guideRepository struct {
	db        *gorm.DB
	redisRepo repository.RedisRepository
}

// NewGuideRepository 创建攻略仓储实例
func NewGuideRepository(db *gorm.DB, redisRepo repository.RedisRepository) GuideRepository {
	return &guideRepository{db: db, redisRepo: redisRepo}
}

// ────────── Redis key 定义 ──────────

const (
	likeCacheTTL = time.Hour
)

func likeStatusKey(strategyID, memberID uint) string {
	return fmt.Sprintf("strategy:like:status:%d:%d", strategyID, memberID)
}
func likeCountKey(strategyID uint) string { return fmt.Sprintf("strategy:likes:count:%d", strategyID) }

func collectStatusKey(strategyID, memberID uint) string {
	return fmt.Sprintf("strategy:collect:status:%d:%d", strategyID, memberID)
}

func cacheStatusValue(ok bool) string {
	if ok {
		return "1"
	}
	return "0"
}

func parseCacheStatus(value string) (bool, bool) {
	switch value {
	case "1":
		return true, true
	case "0":
		return false, true
	default:
		return false, false
	}
}

// Create 创建攻略
func (r *guideRepository) Create(tx *gorm.DB, strategy *entity.Strategy) error {
	return tx.Create(strategy).Error
}

// FindTags 根据标签名查找
func (r *guideRepository) FindTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := tx.Where("tag_name IN ? AND status = 1", tagNames).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// CreateStrategyTags 批量创建攻略-标签关联
func (r *guideRepository) CreateStrategyTags(tx *gorm.DB, strategyID uint, tagIDs []uint) error {
	strategyTags := make([]entity.StrategyTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		strategyTags = append(strategyTags, entity.StrategyTag{
			StrategyID: strategyID,
			TagID:      tagID,
		})
	}
	return tx.Create(&strategyTags).Error
}

// CreateStrategyImages 批量创建攻略图片
func (r *guideRepository) CreateStrategyImages(tx *gorm.DB, strategyID uint, imageURLs []string) error {
	images := make([]entity.StrategyImage, 0, len(imageURLs))
	for i, url := range imageURLs {
		images = append(images, entity.StrategyImage{
			StrategyID: strategyID,
			URL:        url,
			Sort:       i,
		})
	}
	return tx.Create(&images).Error
}

// IncrPublishCount 发布计数 +1
func (r *guideRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}

// FindByID 根据攻略ID查攻略主表（仅上架）
func (r *guideRepository) FindByID(strategyID uint) (*entity.Strategy, error) {
	var strategy entity.Strategy
	if err := r.db.Where("id = ? AND status = 1", strategyID).First(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// FindImagesByStrategyID 查攻略图片列表，按 sort 排序
func (r *guideRepository) FindImagesByStrategyID(strategyID uint) ([]entity.StrategyImage, error) {
	var images []entity.StrategyImage
	if err := r.db.Where("strategy_id = ?", strategyID).
		Order("sort ASC").
		Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

// FindTagsByStrategyID 查攻略标签列表（JOIN strategy_tag + tag，仅启用且未删除的标签）
func (r *guideRepository) FindTagsByStrategyID(strategyID uint) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := r.db.Model(&entity.Tag{}).
		Joins("JOIN strategy_tag ON strategy_tag.tag_id = tag.id").
		Where("strategy_tag.strategy_id = ? AND tag.status = 1 AND tag.delete_time IS NULL", strategyID).
		Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// GetAuthor 获取作者信息
func (r *guideRepository) GetAuthor(memberID uint) (*entity.Member, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// ExistsFollow 检查是否已关注作者
func (r *guideRepository) ExistsFollow(followerID uint, followeeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ? AND status = 1", followerID, followeeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// XAddViewCount 发送浏览量消息到 Redis Stream（异步批量写 DB）
func (r *guideRepository) XAddViewCount(ctx context.Context, strategyID uint) error {
	return r.redisRepo.XAdd(ctx, "stream:content", map[string]interface{}{
		"type": "strategy",
		"id":   strategyID,
	})
}

// IncrViewCount 浏览量 +1
func (r *guideRepository) IncrViewCount(strategyID uint) error {
	return r.db.Model(&entity.Strategy{}).Where("id = ?", strategyID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ExistsLike 检查用户是否已点赞该攻略
func (r *guideRepository) ExistsLike(strategyID uint, memberID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.StrategyLike{}).
		Where("strategy_id = ? AND member_id = ?", strategyID, memberID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateLike 创建点赞记录；仅实际插入成功才同步 like_count +1（并发重复点赞不重复计数）
func (r *guideRepository) CreateLike(tx *gorm.DB, strategyID uint, memberID uint) error {
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.StrategyLike{StrategyID: strategyID, MemberID: memberID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 已点赞过，不重复计数
	}
	return tx.Model(&entity.Strategy{}).Where("id = ?", strategyID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DeleteLike 删除点赞记录（硬删除）；仅实际删除成功才同步 like_count -1（避免并发重复取消把计数减超）
func (r *guideRepository) DeleteLike(tx *gorm.DB, strategyID uint, memberID uint) error {
	res := tx.Unscoped().Where("strategy_id = ? AND member_id = ?", strategyID, memberID).Delete(&entity.StrategyLike{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 本来就没点赞，不减计数
	}
	return tx.Model(&entity.Strategy{}).Where("id = ?", strategyID).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

// GetCachedLikeStatus 读取点赞状态缓存（found=false 表示缓存未命中）
func (r *guideRepository) GetCachedLikeStatus(ctx context.Context, strategyID, memberID uint) (bool, bool, error) {
	var value string
	if err := r.redisRepo.Get(ctx, likeStatusKey(strategyID, memberID), &value); err != nil {
		return false, false, err
	}
	isLiked, ok := parseCacheStatus(value)
	return isLiked, ok, nil
}

// SetCachedLikeStatus 写入点赞状态缓存
func (r *guideRepository) SetCachedLikeStatus(ctx context.Context, strategyID, memberID uint, isLiked bool) {
	_ = r.redisRepo.Set(ctx, likeStatusKey(strategyID, memberID), cacheStatusValue(isLiked), likeCacheTTL)
}

// GetCachedLikeCount 读取点赞数缓存
func (r *guideRepository) GetCachedLikeCount(ctx context.Context, strategyID uint) (int64, error) {
	var count int64
	return count, r.redisRepo.Get(ctx, likeCountKey(strategyID), &count)
}

// SetCachedLikeCount 写入点赞数缓存
func (r *guideRepository) SetCachedLikeCount(ctx context.Context, strategyID uint, count int64) {
	_ = r.redisRepo.Set(ctx, likeCountKey(strategyID), count, likeCacheTTL)
}

// ExistsCollect 检查用户是否已收藏该攻略
func (r *guideRepository) ExistsCollect(strategyID uint, memberID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.StrategyCollect{}).
		Where("strategy_id = ? AND member_id = ?", strategyID, memberID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateCollect 创建收藏记录；仅实际插入成功才同步 collect_count +1（并发重复收藏不重复计数）
func (r *guideRepository) CreateCollect(tx *gorm.DB, strategyID uint, memberID uint) error {
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.StrategyCollect{StrategyID: strategyID, MemberID: memberID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 已收藏过，不重复计数
	}
	// 作者获收藏数 +1（同事务，作者=攻略所属者）
	if err := tx.Model(&entity.Member{}).
		Where("id = (SELECT member_id FROM strategy WHERE id = ?)", strategyID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
		return err
	}
	return tx.Model(&entity.Strategy{}).Where("id = ?", strategyID).
		UpdateColumn("collect_count", gorm.Expr("collect_count + 1")).Error
}

// DeleteCollect 删除收藏记录（硬删除）；仅实际删除成功才同步 collect_count -1（避免并发重复取消把计数减超）
func (r *guideRepository) DeleteCollect(tx *gorm.DB, strategyID uint, memberID uint) error {
	res := tx.Unscoped().Where("strategy_id = ? AND member_id = ?", strategyID, memberID).Delete(&entity.StrategyCollect{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 本来就没收藏，不减计数
	}
	// 作者获收藏数 -1（GREATEST 防减成负数）
	if err := tx.Model(&entity.Member{}).
		Where("id = (SELECT member_id FROM strategy WHERE id = ?)", strategyID).
		UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error; err != nil {
		return err
	}
	return tx.Model(&entity.Strategy{}).Where("id = ?", strategyID).
		UpdateColumn("collect_count", gorm.Expr("collect_count - 1")).Error
}

// GetCachedCollectStatus 读取收藏状态缓存（found=false 表示缓存未命中）
func (r *guideRepository) GetCachedCollectStatus(ctx context.Context, strategyID, memberID uint) (bool, bool, error) {
	var value string
	if err := r.redisRepo.Get(ctx, collectStatusKey(strategyID, memberID), &value); err != nil {
		return false, false, err
	}
	isFavorited, ok := parseCacheStatus(value)
	return isFavorited, ok, nil
}

// SetCachedCollectStatus 写入收藏状态缓存
func (r *guideRepository) SetCachedCollectStatus(ctx context.Context, strategyID, memberID uint, isFavorited bool) {
	_ = r.redisRepo.Set(ctx, collectStatusKey(strategyID, memberID), cacheStatusValue(isFavorited), likeCacheTTL)
}

// UpdateStatus 更新攻略上下架状态；返回是否真正发生了状态变更
// （WHERE status <> ? 保证并发重复删除时第二个请求命中 0 行 → changed=false → 不重复扣减发布数）
func (r *guideRepository) UpdateStatus(tx *gorm.DB, strategyID uint, status int) (bool, error) {
	res := tx.Model(&entity.Strategy{}).Where("id = ? AND status <> ?", strategyID, status).
		UpdateColumn("status", status)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// SoftDelete 软删除攻略：设置 delete_time 为当前时间
func (r *guideRepository) SoftDelete(tx *gorm.DB, strategyID uint) error {
	return tx.Model(&entity.Strategy{}).Where("id = ?", strategyID).
		UpdateColumn("delete_time", time.Now()).Error
}

// UpdateCommentStatusByTarget 批量更新评论状态
func (r *guideRepository) UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error {
	return tx.Model(&entity.Comment{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("status", status).Error
}

// DecrPublishCount 发布计数 -1
func (r *guideRepository) DecrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
}
