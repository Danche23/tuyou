package comment

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// CommentRepository 前台评论仓储接口
type CommentRepository interface {
	FindByID(id uint) (*entity.Comment, error)                             // 查评论（不存在→service 转 404）
	LockCommentByID(tx *gorm.DB, id uint) (*entity.Comment, error)          // 事务内锁评论行（SELECT FOR UPDATE，并发点赞串行化）
	CreateCommentLike(tx *gorm.DB, commentID, memberID uint) error         // 点赞
	DeleteCommentLike(tx *gorm.DB, commentID, memberID uint) (bool, error) // 取消点赞（返回是否真删了）
	IncrLikeCount(tx *gorm.DB, commentID uint) error                       // 点赞数 +1
	DecrLikeCount(tx *gorm.DB, commentID uint) error                       // 点赞数 -1

	CreateComment(tx *gorm.DB, comment *entity.Comment) error                                                      // 发表评论
	FindTargetAuthor(targetType int, targetID uint) (uint, error)                                                  // 校验目标（笔记/攻略）存在，返回作品作者 memberID
	FindTopComments(targetType int, targetID uint, limit int) ([]entity.Comment, error)                            // 一级评论热度最高前 limit 条（热评）
	FindCommentPage(targetType int, targetID uint, excludeIDs []uint, offset, limit int) ([]entity.Comment, error) // 一级评论分页（发布时间倒序，排除热评）
	CountCommentPage(targetType int, targetID uint, excludeIDs []uint) (int64, error)                              // 一级评论（排除热评）总数
	CountRepliesByParentIDs(parentIDs []uint) (map[uint]int64, error)                                              // 批量统计每个一级评论下二级评论总数
	FindTopReplies(parentID uint, limit int) ([]entity.Comment, error)                                             // 某一级评论下热度最高前 limit 条二级评论
	FindReplyPage(parentID uint, excludeIDs []uint, offset, limit int) ([]entity.Comment, error)                   // 二级评论分页（发布时间倒序，排除热评）
	CountReplyPage(parentID uint, excludeIDs []uint) (int64, error)                                                // 二级评论（排除热评）总数
	FindCommentLikes(commentIDs []uint, memberID uint) (map[uint]bool, error)                                      // 批量查当前用户点赞状态
	FindMembersByIDs(memberIDs []uint) (map[uint]*entity.Member, error)                                            // 批量查用户信息
	DeleteCommentByID(tx *gorm.DB, id uint) error                                                                  // 软删除评论
	DeleteRepliesByParentID(tx *gorm.DB, parentID uint) error                                                      // 软删除某一级评论下所有二级评论

	CountRepliesByParentID(tx *gorm.DB, parentID uint) (int64, error)                 // 统计单个一级评论下未删二级数（事务内）
	IncrTargetCommentCount(tx *gorm.DB, targetType int, targetID uint) error          // 目标 note/strategy 评论数 +1（SQL 原子）
	DecrTargetCommentCount(tx *gorm.DB, targetType int, targetID uint, n int64) error // 目标评论数 -n（GREATEST 防负）
	RecomputeHeat(tx *gorm.DB, commentID uint) error                                  // 事务内重算单条评论热度（4.4 公式）
}
