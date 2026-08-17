package comment

// 前台评论模块 —— 仓储层接口
//
// 分层思想：Service 层只面向这个接口编程（依赖接口，不依赖实现），
// 好处是：
//   - Service 的单元测试可以拿一个 mock 实现替换掉真实的 MySQL（见 service 的测试）。
//   - 换数据库 / 改 SQL 只动 repository.go，Service 一行不改。
//
// 方法分类看名字就能懂：
//   - Find/Count 开头：各种查询（评论、用户、目标、点赞态）
//   - Create/Delete/Incr/Decr 开头：写操作，几乎都带 tx *gorm.DB（事务对象）
//
// 带 tx 参数 = 这个方法必须在调用方开的"事务里"执行，由 service 用 s.db.Transaction 包起来；
// 不带 tx = 独立查询，直接用仓储自己的 db。这是本项目约定好的规矩。
import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// CommentRepository 前台评论仓储接口
type CommentRepository interface {
	FindByID(id uint) (*entity.Comment, error)                             // 查评论（不存在→service 转 404）
	LockCommentByID(tx *gorm.DB, id uint) (*entity.Comment, error)         // 事务内锁评论行（SELECT FOR UPDATE，并发点赞串行化）
	CreateCommentLike(tx *gorm.DB, commentID, memberID uint) error         // 点赞
	DeleteCommentLike(tx *gorm.DB, commentID, memberID uint) (bool, error) // 取消点赞（返回是否真删了）
	IncrLikeCount(tx *gorm.DB, commentID uint) error                       // 点赞数 +1
	DecrLikeCount(tx *gorm.DB, commentID uint) error                       // 点赞数 -1

	CreateComment(tx *gorm.DB, comment *entity.Comment) error                                                      // 发表评论
	FindTargetAuthor(targetType int, targetID uint) (uint, string, error)                                          // 校验目标（笔记/攻略）存在，返回 (作品作者 memberID, 标题 title)
	FindTopComments(targetType int, targetID uint, limit int) ([]entity.Comment, error)                            // 一级评论热度最高前 limit 条（热评）
	FindCommentPage(targetType int, targetID uint, excludeIDs []uint, offset, limit int) ([]entity.Comment, error) // 一级评论分页（发布时间倒序，排除热评）
	CountCommentPage(targetType int, targetID uint, excludeIDs []uint) (int64, error)                              // 一级评论（排除热评）总数
	CountAllComments(targetType int, targetID uint) (int64, error)                                                 // 目标下全量评论数（一级+二级，未删）
	CountRepliesByParentIDs(parentIDs []uint) (map[uint]int64, error)                                              // 批量统计每个一级评论下二级评论总数
	FindTopReplies(parentID uint, limit int) ([]entity.Comment, error)                                             // 某一级评论下热度最高前 limit 条二级评论
	FindRepliesByParentIDs(parentIDs []uint) ([]entity.Comment, error)                                             // 批量查询多个一级评论下的全部二级评论（C4 防 N+1）
	FindReplyPage(parentID uint, excludeIDs []uint, offset, limit int) ([]entity.Comment, error)                   // 二级评论分页（发布时间正序，排除热评）
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
