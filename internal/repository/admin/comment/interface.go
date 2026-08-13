package comment

// 后台评论管理模块 —— 仓储层接口
//
// 与前台仓储最大的区别在"软删"的处理：
//   - 前台仓储默认作用域：自动过滤软删（delete_time IS NULL），用户看不到已删评论。
//   - 后台仓储全用 Unscoped()：把软删的也查出来，因为管理员要能看到并清理"已删除"的评论。
//
// 另外一个独有职责：物理删除。Delete / BatchDelete 是事务内"真删"，
// 删除时要把三样东西一起清干净：
//  1. 评论行本身（一级还要连带其下所有二级）
//  2. 这些评论的点赞记录（comment_like 表）
//  3. 目标 note/strategy 的 comment_count 相应扣减
//
// 查询方法带 filters map（拼装筛选条件），排序字段由 Service 白名单传入，防止 SQL 注入。
import (
	"context"

	"tuyou/internal/model/entity"
)

// CommentRepository 后台评论管理仓储接口
type CommentRepository interface {
	// Count 统计评论数（Unscoped 含软删记录）
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	// FindList 分页查询评论列表（Unscoped 含软删记录）；sortField 来自 Service 白名单，空=默认 id 升序
	FindList(ctx context.Context, filters map[string]interface{}, sortField, sortOrder string, page, size int) ([]entity.Comment, error)
	// FindByID 查单条评论（Unscoped 含软删记录，不存在→nil, nil）
	FindByID(ctx context.Context, id uint) (*entity.Comment, error)
	// FindCommentsByIDs 批量查询评论（Unscoped 含软删记录）
	FindCommentsByIDs(ctx context.Context, ids []uint) ([]entity.Comment, error)
	// FindMembersByIDs 批量查询用户信息
	FindMembersByIDs(ctx context.Context, memberIDs []uint) (map[uint]*entity.Member, error)
	// FindNoteTitles 批量查询笔记标题 map[noteID]title
	FindNoteTitles(ctx context.Context, ids []uint) (map[uint]string, error)
	// FindStrategyTitles 批量查询攻略标题 map[strategyID]title
	FindStrategyTitles(ctx context.Context, ids []uint) (map[uint]string, error)
	// Delete 事务内硬删一条评论（一级连带删其下所有二级 + 点赞记录），并按删除行数扣减目标 comment_count
	Delete(ctx context.Context, comment *entity.Comment) error
	// BatchDelete 事务内批量硬删评论（不存在的 ID 已在查询阶段跳过）
	BatchDelete(ctx context.Context, comments []entity.Comment) error
}
