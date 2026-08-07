package guide

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

type guideRepository struct {
	db *gorm.DB
}

// NewGuideRepository 创建后台攻略仓储实例
func NewGuideRepository(db *gorm.DB) GuideRepository {
	return &guideRepository{db: db}
}

// applyFilters 拼装动态查询条件（针对 strategy 表）
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["title"]; ok {
		query = query.Where("title LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["category"]; ok {
		query = query.Where("category = ?", v)
	}
	if v, ok := filters["difficulty"]; ok {
		query = query.Where("difficulty = ?", v)
	}
	if v, ok := filters["is_slide"]; ok {
		query = query.Where("is_slide = ?", v)
	}
	if v, ok := filters["begin_time"]; ok {
		query = query.Where("publish_time >= ?", v)
	}
	if v, ok := filters["finish_time"]; ok {
		query = query.Where("publish_time <= ?", v)
	}
	return query
}

// Count 统计攻略数量
func (r *guideRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Strategy{})
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindList 分页查询攻略列表
func (r *guideRepository) FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Strategy, error) {
	strategies := make([]entity.Strategy, 0)

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Strategy{})
	query = applyFilters(query, filters)

	// 排序（sortField 来自 Service 白名单，安全）
	if sortField != "" {
		query = query.Order(sortField + " " + sortOrder)
	} else {
		query = query.Order("id DESC")
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

// FindByID 根据 ID 查询单篇攻略（Unscoped 含软删除记录，不存在返回 nil, nil）
func (r *guideRepository) FindByID(ctx context.Context, id uint) (*entity.Strategy, error) {
	var strategy entity.Strategy
	err := r.db.WithContext(ctx).Unscoped().First(&strategy, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &strategy, nil
}

// FindTagsByStrategyIDs 批量查询攻略标签 map[strategyID][]tagName
func (r *guideRepository) FindTagsByStrategyIDs(ctx context.Context, strategyIDs []uint) (map[uint][]string, error) {
	if len(strategyIDs) == 0 {
		return map[uint][]string{}, nil
	}

	var rows []struct {
		StrategyID uint   `gorm:"column:strategy_id"`
		TagName    string `gorm:"column:tag_name"`
	}
	if err := r.db.WithContext(ctx).Table("strategy_tag").
		Select("strategy_tag.strategy_id AS strategy_id, tag.tag_name AS tag_name").
		Joins("JOIN tag ON tag.id = strategy_tag.tag_id").
		Where("strategy_tag.strategy_id IN ?", strategyIDs).
		Order("strategy_tag.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	tagsMap := make(map[uint][]string)
	for _, row := range rows {
		tagsMap[row.StrategyID] = append(tagsMap[row.StrategyID], row.TagName)
	}
	return tagsMap, nil
}

// FindMembersByIDs 批量查询作者信息
func (r *guideRepository) FindMembersByIDs(ctx context.Context, memberIDs []uint) ([]entity.Member, error) {
	if len(memberIDs) == 0 {
		return []entity.Member{}, nil
	}

	members := make([]entity.Member, 0)
	if err := r.db.WithContext(ctx).Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// FindCoverImagesByStrategyIDs 批量查询攻略封面图（每篇第一张）map[strategyID]url
func (r *guideRepository) FindCoverImagesByStrategyIDs(ctx context.Context, strategyIDs []uint) (map[uint]string, error) {
	if len(strategyIDs) == 0 {
		return map[uint]string{}, nil
	}

	var rows []struct {
		StrategyID uint   `gorm:"column:strategy_id"`
		URL        string `gorm:"column:url"`
	}
	if err := r.db.WithContext(ctx).Table("strategy_image").
		Select("strategy_id AS strategy_id, url AS url").
		Where("strategy_id IN ?", strategyIDs).
		Order("sort ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 行已按 sort 升序排列，每篇攻略取第一次出现的那张（即 sort 最小）
	covers := make(map[uint]string)
	seen := make(map[uint]bool)
	for _, row := range rows {
		if seen[row.StrategyID] {
			continue
		}
		seen[row.StrategyID] = true
		covers[row.StrategyID] = row.URL
	}
	return covers, nil
}

// Update 更新攻略字段（Unscoped 以支持恢复已软删攻略）
func (r *guideRepository) Update(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Unscoped().Model(&entity.Strategy{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateCommentStatusByTarget 更新目标攻略/笔记下的评论状态（2=攻略）
func (r *guideRepository) UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error {
	return tx.Model(&entity.Comment{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("status", status).Error
}

// IncrPublishCount 作者发布数 +1
func (r *guideRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}

// DecrPublishCount 作者发布数 -1
func (r *guideRepository) DecrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
}

// deleteStrategyInTx 事务内删除单篇攻略及其关联数据；decrPublish=true 时作者发布数 -1
func deleteStrategyInTx(tx *gorm.DB, strategyID uint, memberID uint, decrPublish bool) error {
	// 关联数据：图片 / 标签 / 点赞 / 收藏
	if err := tx.Unscoped().Where("strategy_id = ?", strategyID).Delete(&entity.StrategyImage{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("strategy_id = ?", strategyID).Delete(&entity.StrategyTag{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("strategy_id = ?", strategyID).Delete(&entity.StrategyLike{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("strategy_id = ?", strategyID).Delete(&entity.StrategyCollect{}).Error; err != nil {
		return err
	}
	// 评论点赞（先删，子查询引用 comment 表）
	if err := tx.Unscoped().
		Where("comment_id IN (SELECT id FROM comment WHERE target_type = ? AND target_id = ?)", 2, strategyID).
		Delete(&entity.CommentLike{}).Error; err != nil {
		return err
	}
	// 评论（攻略 target_type=2）
	if err := tx.Unscoped().Where("target_type = ? AND target_id = ?", 2, strategyID).Delete(&entity.Comment{}).Error; err != nil {
		return err
	}
	// 主行
	if err := tx.Unscoped().Delete(&entity.Strategy{}, strategyID).Error; err != nil {
		return err
	}
	// 仍上架的攻略：作者发布数 -1
	if decrPublish {
		return tx.Model(&entity.Member{}).Where("id = ?", memberID).
			UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
	}
	return nil
}

// Delete 硬删除攻略及其关联数据（事务）；decrPublish=true 时作者发布数 -1
func (r *guideRepository) Delete(ctx context.Context, strategyID uint, memberID uint, decrPublish bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteStrategyInTx(tx, strategyID, memberID, decrPublish)
	})
}

// FindStrategiesByIDs 根据 ID 列表批量查询攻略（Unscoped 含软删，不存在的 ID 自动跳过）
func (r *guideRepository) FindStrategiesByIDs(ctx context.Context, ids []uint) ([]entity.Strategy, error) {
	if len(ids) == 0 {
		return []entity.Strategy{}, nil
	}

	strategies := make([]entity.Strategy, 0)
	if err := r.db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Order("id DESC").Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

// BatchDelete 批量硬删除攻略及其关联数据（单事务）；仍上架的攻略作者发布数 -1
func (r *guideRepository) BatchDelete(ctx context.Context, strategies []entity.Strategy) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range strategies {
			g := &strategies[i]
			if err := deleteStrategyInTx(tx, g.ID, g.MemberID, g.Status == 1); err != nil {
				return err
			}
		}
		return nil
	})
}
