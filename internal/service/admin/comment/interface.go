package comment

// 后台评论管理模块 —— 服务层接口
//
// 4 个能力，对应 postman 文档的 4 个接口：
//   - ListComments      分页列表：多条件筛选 + 排序 + 分页，并组装出
//     target_title / reply_to_nickname / level / status 等"需要跨表才知道"的字段。
//   - GetCommentDetail  单条详情（复用列表的组装逻辑）。
//   - DeleteComment     彻底删除一条（一级连带其下所有二级 + 点赞）。
//   - BatchDeleteComments 批量彻底删除（不存在的 ID 自动跳过）。
//
// 与前台服务的职责对比：前台做业务规则（谁能删、热评置顶、两级模型），
// 后台更接近"数据管理"（筛选、看状态、清理数据），业务规则较少。
import (
	"context"

	request "tuyou/internal/model/dto/request/admin/comment"
	response "tuyou/internal/model/dto/response/admin/comment"
)

// CommentService 后台评论管理服务接口
type CommentService interface {
	// ListComments 分页查询评论列表
	ListComments(ctx context.Context, req *request.ListCommentReq) (*response.ListCommentResp, error)
	// GetCommentDetail 获取评论详情
	GetCommentDetail(ctx context.Context, id uint) (*response.CommentItem, error)
	// DeleteComment 彻底删除评论（一级连带删除其下所有二级）
	DeleteComment(ctx context.Context, id uint) error
	// BatchDeleteComments 批量彻底删除评论（不存在的 ID 自动跳过）
	BatchDeleteComments(ctx context.Context, req *request.BatchDeleteReq) error
}
