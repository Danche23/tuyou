package comment

// 前台评论模块 —— 仓储层实现（真正写 SQL 的地方）
//
// 这一层就是"把 Go 结构体翻译成 SQL"，几个贯穿全文件的约定：
//   - 默认作用域：不带 Unscoped() 的查询会自动过滤掉软删记录
//     （delete_time IS NULL，由 GORM 自动追加），所以前台"看不见已删除的评论"。
//   - 软删除：DeleteCommentByID 用 tx.Delete(...)，本质是 UPDATE 置 delete_time，
//     评论行还在，只是被藏起来。
//   - 原子加减：comment_count / like_count 的增减用 gorm.Expr 拼 SQL 表达式，
//     在数据库端原子执行，避免"先查后改"在多并发下互相覆盖。
//   - 批量查询：Find*ByIDs 都一次性 IN 查回 map，避免 N+1 查询。
//   - RecomputeHeatExpr：评论热度公式，见下方常量的注释。
//
// 前台（web）与后台（admin）各有一个评论仓储：前台默认"隐藏软删"，后台用 Unscoped() "连软删一起看"。
import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// FindByID 按 ID 查评论（软删的评论查不到，返回错误由 service 转 404）
func (r *commentRepository) FindByID(id uint) (*entity.Comment, error) {
	var comment entity.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// LockCommentByID 事务内锁定评论行（SELECT ... FOR UPDATE），串行化并发点赞。
// 不锁行时，"删点赞记录 + 插点赞记录"在并发下会撞唯一索引 idx_comment_member 导致 500。
func (r *commentRepository) LockCommentByID(tx *gorm.DB, id uint) (*entity.Comment, error) {
	var comment entity.Comment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// CreateCommentLike 点赞：写入 comment_like 记录（调用方需先 DeleteCommentLike，保证不重复）
func (r *commentRepository) CreateCommentLike(tx *gorm.DB, commentID, memberID uint) error {
	return tx.Create(&entity.CommentLike{
		CommentID: commentID,
		MemberID:  memberID,
	}).Error
}

// DeleteCommentLike 取消点赞：硬删除，返回是否真的删掉了（true=之前已点赞）
func (r *commentRepository) DeleteCommentLike(tx *gorm.DB, commentID, memberID uint) (bool, error) {
	res := tx.Unscoped().
		Where("comment_id = ? AND member_id = ?", commentID, memberID).
		Delete(&entity.CommentLike{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// IncrLikeCount 评论点赞数 +1
func (r *commentRepository) IncrLikeCount(tx *gorm.DB, commentID uint) error {
	return tx.Model(&entity.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DecrLikeCount 评论点赞数 -1（GREATEST 防负，与 DecrTargetCommentCount 保持一致）
func (r *commentRepository) DecrLikeCount(tx *gorm.DB, commentID uint) error {
	return tx.Model(&entity.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

// CreateComment 发表评论
func (r *commentRepository) CreateComment(tx *gorm.DB, comment *entity.Comment) error {
	return tx.Create(comment).Error
}

// FindTargetAuthor 校验目标（笔记/攻略）存在，返回作品作者 memberID；不存在（含软删）→ 返回错误
func (r *commentRepository) FindTargetAuthor(targetType int, targetID uint) (uint, error) {
	if targetType == 1 {
		var note entity.Note
		if err := r.db.Select("member_id").First(&note, targetID).Error; err != nil {
			return 0, err
		}
		return note.MemberID, nil
	}
	var st entity.Strategy
	if err := r.db.Select("member_id").First(&st, targetID).Error; err != nil {
		return 0, err
	}
	return st.MemberID, nil
}

// FindTopComments 一级评论热度最高前 limit 条（热评，不占分页计数）
func (r *commentRepository) FindTopComments(targetType int, targetID uint, limit int) ([]entity.Comment, error) {
	var comments []entity.Comment
	if err := r.db.Where("target_type = ? AND target_id = ? AND reply_to_id IS NULL", targetType, targetID).
		Order("heat DESC, id ASC").Limit(limit).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// FindCommentPage 一级评论分页（发布时间倒序，排除热评）
func (r *commentRepository) FindCommentPage(targetType int, targetID uint, excludeIDs []uint, offset, limit int) ([]entity.Comment, error) {
	var comments []entity.Comment
	query := r.db.Where("target_type = ? AND target_id = ? AND reply_to_id IS NULL", targetType, targetID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Order("create_time DESC, id DESC").Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// CountCommentPage 一级评论（排除热评）总数
func (r *commentRepository) CountCommentPage(targetType int, targetID uint, excludeIDs []uint) (int64, error) {
	var total int64
	query := r.db.Model(&entity.Comment{}).Where("target_type = ? AND target_id = ? AND reply_to_id IS NULL", targetType, targetID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountAllComments 统计目标下全量评论数（一级+二级，delete_time IS NULL 由 GORM 自动追加）
// 用于评论列表响应里的 totalComments：实时统计、天然自愈，不依赖 ±1 维护的计数
func (r *commentRepository) CountAllComments(targetType int, targetID uint) (int64, error) {
	var total int64
	if err := r.db.Model(&entity.Comment{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountRepliesByParentIDs 批量统计每个一级评论下的二级评论总数
func (r *commentRepository) CountRepliesByParentIDs(parentIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(parentIDs))
	if len(parentIDs) == 0 {
		return result, nil
	}
	var counts []struct {
		ReplyToID uint  `gorm:"column:reply_to_id"`
		Cnt       int64 `gorm:"column:cnt"`
	}
	if err := r.db.Model(&entity.Comment{}).
		Select("reply_to_id, COUNT(*) AS cnt").
		Where("reply_to_id IN ?", parentIDs).
		Group("reply_to_id").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	for _, c := range counts {
		result[c.ReplyToID] = c.Cnt
	}
	return result, nil
}

// FindTopReplies 某一级评论下热度最高前 limit 条二级评论
func (r *commentRepository) FindTopReplies(parentID uint, limit int) ([]entity.Comment, error) {
	var comments []entity.Comment
	if err := r.db.Where("reply_to_id = ?", parentID).
		Order("heat DESC, id ASC").Limit(limit).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// FindReplyPage 二级评论分页（发布时间倒序，排除热评）
func (r *commentRepository) FindReplyPage(parentID uint, excludeIDs []uint, offset, limit int) ([]entity.Comment, error) {
	var comments []entity.Comment
	query := r.db.Where("reply_to_id = ?", parentID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Order("create_time DESC, id DESC").Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// CountReplyPage 二级评论（排除热评）总数
func (r *commentRepository) CountReplyPage(parentID uint, excludeIDs []uint) (int64, error) {
	var total int64
	query := r.db.Model(&entity.Comment{}).Where("reply_to_id = ?", parentID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindCommentLikes 批量查当前用户对哪些评论点过赞
func (r *commentRepository) FindCommentLikes(commentIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(commentIDs))
	if len(commentIDs) == 0 {
		return result, nil
	}
	var likes []entity.CommentLike
	if err := r.db.Where("comment_id IN ? AND member_id = ?", commentIDs, memberID).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, l := range likes {
		result[l.CommentID] = true
	}
	return result, nil
}

// FindMembersByIDs 批量查用户信息
func (r *commentRepository) FindMembersByIDs(memberIDs []uint) (map[uint]*entity.Member, error) {
	result := make(map[uint]*entity.Member, len(memberIDs))
	if len(memberIDs) == 0 {
		return result, nil
	}
	var members []entity.Member
	if err := r.db.Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	for i := range members {
		result[members[i].ID] = &members[i]
	}
	return result, nil
}

// DeleteCommentByID 软删除评论（置 delete_time）
func (r *commentRepository) DeleteCommentByID(tx *gorm.DB, id uint) error {
	return tx.Delete(&entity.Comment{}, id).Error
}

// DeleteRepliesByParentID 软删除某一级评论下所有二级评论
func (r *commentRepository) DeleteRepliesByParentID(tx *gorm.DB, parentID uint) error {
	return tx.Where("reply_to_id = ?", parentID).Delete(&entity.Comment{}).Error
}

// CountRepliesByParentID 统计单个一级评论下未删二级评论数（事务内）
func (r *commentRepository) CountRepliesByParentID(tx *gorm.DB, parentID uint) (int64, error) {
	var total int64
	if err := tx.Model(&entity.Comment{}).Where("reply_to_id = ?", parentID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// targetModel targetType → 目标表实体（1=笔记，2=攻略）
func targetModel(targetType int) interface{} {
	if targetType == 1 {
		return &entity.Note{}
	}
	return &entity.Strategy{}
}

// IncrTargetCommentCount 目标 note/strategy 的 comment_count +1（SQL 原子更新）
func (r *commentRepository) IncrTargetCommentCount(tx *gorm.DB, targetType int, targetID uint) error {
	return tx.Model(targetModel(targetType)).
		Where("id = ?", targetID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// DecrTargetCommentCount 目标 note/strategy 的 comment_count -n（GREATEST 防负，事务内）
func (r *commentRepository) DecrTargetCommentCount(tx *gorm.DB, targetType int, targetID uint, n int64) error {
	return tx.Model(targetModel(targetType)).
		Where("id = ?", targetID).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", n)).Error
}

// RecomputeHeatExpr 评论热度公式 SQL 表达式（设计文档 4.4）：
// heat = 互动分 × 时间因子 + 新鲜度加分
//
//	互动分   = like_count + reply_count × 2
//	时间因子 = 28天内=1；之后每 7 天 ×0.6
//	新鲜度   = 发布后第 1/2/3 周 +6/+4/+2，之后 +0
//
// reply_count：一级评论 = 其下未删二级数（用双层派生表规避 MySQL 1093：不能在同一 UPDATE 里子查询被更新的表），二级评论 = 0（子查询自然不命中）
const RecomputeHeatExpr = `
	(like_count + COALESCE((
		SELECT cnt FROM (
			SELECT reply_to_id, COUNT(*) AS cnt
			FROM comment
			WHERE delete_time IS NULL
			GROUP BY reply_to_id
		) t WHERE t.reply_to_id = comment.id
	), 0) * 2)
	* POWER(0.6, FLOOR(GREATEST(DATEDIFF(NOW(), create_time) - 28, 0) / 7))
	+ CASE
		WHEN DATEDIFF(NOW(), create_time) < 7 THEN 6
		WHEN DATEDIFF(NOW(), create_time) < 14 THEN 4
		WHEN DATEDIFF(NOW(), create_time) < 21 THEN 2
		ELSE 0
	END`

// RecomputeHeat 事务内按公式重算单条评论热度（写操作后调用：发评论/删除/点赞）
func (r *commentRepository) RecomputeHeat(tx *gorm.DB, commentID uint) error {
	return tx.Model(&entity.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("heat", gorm.Expr(RecomputeHeatExpr)).Error
}
