package comment

// 后台评论管理模块 —— 仓储层实现（SQL 都在这）
//
// 三个要点：
//  1. 查询全部 Unscoped()：连软删记录一起看，status 由 service 层从 delete_time 派生。
//  2. applyFilters：把 service 传来的筛选条件翻译成 WHERE 子句。
//     昵称/标题用 IN (SELECT ... LIKE) 子查询做模糊匹配；
//     时间范围用 create_time >= / <=；层级用 reply_to_id IS NULL / IS NOT NULL。
//  3. deleteCommentInTx：后台删除的核心，见函数注释——物理删 + 级联删二级 +
//     清点赞 + 扣 comment_count，且只按"仍正常的行"扣，避免对已软删评论重复扣减。
//
// 与前台仓储的对应关系：
//   前台 DeleteCommentByID（软删，藏起来）←→ 后台 Delete/BatchDelete（硬删，抹掉）。
//   前台 DecrTargetCommentCount ←→ 后台同样的扣减逻辑，但后台要在"删完"后扣。
import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"

	repoComment "tuyou/internal/repository/web/comment" // 复用前台导出的热度公式常量 RecomputeHeatExpr（单一来源，避免公式两处漂移）
)

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建后台评论仓储实例
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// applyFilters 拼装动态查询条件（针对 comment 表）
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["target_type"]; ok {
		query = query.Where("target_type = ?", v)
	}
	if v, ok := filters["nickname"]; ok {
		query = query.Where("member_id IN (SELECT id FROM member WHERE nickname LIKE ?)", "%"+v.(string)+"%")
	}
	if v, ok := filters["title"]; ok {
		t := "%" + v.(string) + "%"
		query = query.Where(
			"(target_type = 1 AND target_id IN (SELECT id FROM note WHERE title LIKE ?)) OR (target_type = 2 AND target_id IN (SELECT id FROM strategy WHERE title LIKE ?))",
			t, t,
		)
	}
	if v, ok := filters["begin_time"]; ok {
		query = query.Where("create_time >= ?", v)
	}
	if v, ok := filters["finish_time"]; ok {
		query = query.Where("create_time <= ?", v)
	}
	if v, ok := filters["level"]; ok {
		if v.(string) == "一级" {
			query = query.Where("reply_to_id IS NULL")
		} else {
			query = query.Where("reply_to_id IS NOT NULL")
		}
	}
	return query
}

// Count 统计评论数（Unscoped 含软删记录，后台要看到"已删除"状态）
func (r *commentRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Comment{})
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindList 分页查询评论列表（Unscoped 含软删记录）
func (r *commentRepository) FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Comment, error) {
	comments := make([]entity.Comment, 0)

	query := r.db.WithContext(ctx).Unscoped().Model(&entity.Comment{})
	query = applyFilters(query, filters)

	// 排序（sortField 来自 Service 白名单，安全；不传默认 id 升序）
	if sortField != "" {
		query = query.Order(sortField + " " + sortOrder)
	} else {
		query = query.Order("id ASC")
	}

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// FindByID 按 ID 查询评论（Unscoped 含软删记录，不存在→nil, nil）
func (r *commentRepository) FindByID(ctx context.Context, id uint) (*entity.Comment, error) {
	var comment entity.Comment
	err := r.db.WithContext(ctx).Unscoped().First(&comment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// FindCommentsByIDs 按 ID 列表批量查询评论（Unscoped 含软删记录）
func (r *commentRepository) FindCommentsByIDs(ctx context.Context, ids []uint) ([]entity.Comment, error) {
	if len(ids) == 0 {
		return []entity.Comment{}, nil
	}
	comments := make([]entity.Comment, 0)
	if err := r.db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Order("id ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// FindMembersByIDs 批量查询用户信息
func (r *commentRepository) FindMembersByIDs(ctx context.Context, memberIDs []uint) (map[uint]*entity.Member, error) {
	result := make(map[uint]*entity.Member, len(memberIDs))
	if len(memberIDs) == 0 {
		return result, nil
	}
	var members []entity.Member
	if err := r.db.WithContext(ctx).Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	for i := range members {
		result[members[i].ID] = &members[i]
	}
	return result, nil
}

// FindNoteTitles 批量查询笔记标题 map[noteID]title
func (r *commentRepository) FindNoteTitles(ctx context.Context, ids []uint) (map[uint]string, error) {
	result := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var rows []struct {
		ID    uint   `gorm:"column:id"`
		Title string `gorm:"column:title"`
	}
	if err := r.db.WithContext(ctx).Table("note").Select("id, title").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row.Title
	}
	return result, nil
}

// FindStrategyTitles 批量查询攻略标题 map[strategyID]title
func (r *commentRepository) FindStrategyTitles(ctx context.Context, ids []uint) (map[uint]string, error) {
	result := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var rows []struct {
		ID    uint   `gorm:"column:id"`
		Title string `gorm:"column:title"`
	}
	if err := r.db.WithContext(ctx).Table("strategy").Select("id, title").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row.Title
	}
	return result, nil
}

// targetModel targetType → 目标表实体（1=笔记，2=攻略）
func targetModel(targetType int) interface{} {
	if targetType == 1 {
		return &entity.Note{}
	}
	return &entity.Strategy{}
}

// deleteCommentInTx 事务内硬删一条评论：
// 物理删自身 +（若一级）其下所有二级（含已软删的）+ 这些评论的点赞记录，
// 并按"仍正常"的行数扣减目标 comment_count（已软删行软删时已扣过，不再重复扣）。
// 若删除的是二级评论：其归属一级评论的回复数变了，还要立即重算它的热度，避免等小时任务兜底。
// 若自身已不存在（批量场景下已被父评论连带删除）则直接跳过，避免重复扣减。
func deleteCommentInTx(tx *gorm.DB, comment *entity.Comment) error {
	// 0. 先确认自身还存在
	var exists int64
	if err := tx.Model(&entity.Comment{}).Unscoped().Where("id = ?", comment.ID).Count(&exists).Error; err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}

	// 1. 删除行数（用于 comment_count 扣减）：只统计"当前仍正常"的行。
	// 已软删的行在软删那一刻已扣减过 comment_count，硬删时不能再扣（否则过度扣减）。
	// 二级=0或1；一级=0或1 + 其下未删二级数（默认作用域自动过滤 delete_time IS NULL）。
	var n int64 = 0
	if !comment.DeletedAt.Valid {
		n = 1
	}
	if comment.ReplyToID == nil {
		var cnt int64
		if err := tx.Model(&entity.Comment{}).Where("reply_to_id = ?", comment.ID).Count(&cnt).Error; err != nil {
			return err
		}
		n += cnt
	}

	// 2. 物理删这些评论的点赞记录（自身 + 若一级则其下二级；先删，避免子查询引用已删除的 comment 行）
	if err := tx.Unscoped().
		Where("comment_id IN (SELECT id FROM comment WHERE id = ? OR reply_to_id = ?)", comment.ID, comment.ID).
		Delete(&entity.CommentLike{}).Error; err != nil {
		return err
	}
	// 3. 若是一级评论：物理删其下所有二级评论
	if comment.ReplyToID == nil {
		if err := tx.Unscoped().Where("reply_to_id = ?", comment.ID).Delete(&entity.Comment{}).Error; err != nil {
			return err
		}
	}
	// 4. 物理删自身
	if err := tx.Unscoped().Delete(&entity.Comment{}, comment.ID).Error; err != nil {
		return err
	}
	// 5. 目标 note/strategy comment_count -n（GREATEST 防负）
	if err := tx.Model(targetModel(comment.TargetType)).
		Where("id = ?", comment.TargetID).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", n)).Error; err != nil {
		return err
	}

	// 6. 若删除的是二级评论：其归属一级评论的回复数变了，立即重算它的热度。
	// 复用前台 RecomputeHeatExpr（同一公式）；父级若已软删（delete_time 非空），
	// 默认作用域自动不命中，热度也不该展示，等同跳过。
	if comment.ReplyToID != nil {
		return tx.Model(&entity.Comment{}).
			Where("id = ?", *comment.ReplyToID).
			UpdateColumn("heat", gorm.Expr(repoComment.RecomputeHeatExpr)).Error
	}
	return nil
}

// Delete 事务内硬删单条评论
func (r *commentRepository) Delete(ctx context.Context, comment *entity.Comment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteCommentInTx(tx, comment)
	})
}

// BatchDelete 事务内批量硬删评论（不存在的 ID 已在查询阶段跳过）
func (r *commentRepository) BatchDelete(ctx context.Context, comments []entity.Comment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range comments {
			if err := deleteCommentInTx(tx, &comments[i]); err != nil {
				return err
			}
		}
		return nil
	})
}
