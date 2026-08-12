package comment

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// CommentRepository 前台评论仓储接口
type CommentRepository interface {
	FindByID(id uint) (*entity.Comment, error)                             // 查评论（不存在→service 转 404）
	CreateCommentLike(tx *gorm.DB, commentID, memberID uint) error         // 点赞
	DeleteCommentLike(tx *gorm.DB, commentID, memberID uint) (bool, error) // 取消点赞（返回是否真删了）
	IncrLikeCount(tx *gorm.DB, commentID uint) error                       // 点赞数 +1
	DecrLikeCount(tx *gorm.DB, commentID uint) error                       // 点赞数 -1
}
