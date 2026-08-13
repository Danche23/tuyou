package comment

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
