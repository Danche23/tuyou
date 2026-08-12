package comment

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
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

// DecrLikeCount 评论点赞数 -1
func (r *commentRepository) DecrLikeCount(tx *gorm.DB, commentID uint) error {
	return tx.Model(&entity.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}
