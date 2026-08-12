package comment

import (
	responseComment "tuyou/internal/model/dto/response/web/comment"
)

// CommentService 前台评论服务接口
type CommentService interface {
	ToggleLike(commentID uint, memberID uint) (*responseComment.LikeToggleResponse, error) // 点赞/取消点赞
}
