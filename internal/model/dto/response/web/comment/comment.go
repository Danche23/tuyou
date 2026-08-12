package comment

// LikeToggleResponse 点赞/取消点赞响应
type LikeToggleResponse struct {
	IsLiked   bool  `json:"isLiked"`
	LikeCount int64 `json:"likeCount"`
}
