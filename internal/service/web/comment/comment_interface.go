package comment

// 前台评论模块 —— 服务层接口
//
// Service 是"业务规则"的家，本模块的 5 个能力：
//   - ToggleLike    点赞/取消点赞（翻转）
//   - CreateComment 发表评论（一级或二级）
//   - ListComments  一级评论列表（热评置顶 + 分页）
//   - DeleteComment 删除评论（软删，一级连带二级）
//   - ListReplies   二级评论列表（热回复置顶 + 分页）
//
// 与 Repository 接口对应：Service 接口是给 Controller 用的，Repository 接口是 Service 自己用的。
// 都返回 (数据, error)，error 统一是 *errors.BizError（带业务错误码，见 pkg/errors）。
import (
	requestComment "tuyou/internal/model/dto/request/web/comment"
	responseComment "tuyou/internal/model/dto/response/web/comment"
	"tuyou/pkg/response"
)

// CommentService 前台评论服务接口
type CommentService interface {
	ToggleLike(commentID uint, memberID uint) (*responseComment.LikeToggleResponse, error)                       // 点赞/取消点赞
	CreateComment(memberID uint, req *requestComment.CreateCommentRequest) (*responseComment.CommentItem, error) // 发表评论
	ListComments(memberID uint, req *requestComment.ListCommentRequest) (*response.PageResponse, error)          // 评论列表
	DeleteComment(commentID uint, memberID uint) error                                                           // 删除评论
	ListReplies(commentID uint, memberID uint, page, size int) (*response.PageResponse, error)                   // 二级评论列表
}
