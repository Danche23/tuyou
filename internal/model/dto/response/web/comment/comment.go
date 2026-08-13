package comment

// LikeToggleResponse 点赞/取消点赞响应
type LikeToggleResponse struct {
	IsLiked   bool  `json:"isLiked"`
	LikeCount int64 `json:"likeCount"`
}

// CommentUser 评论用户信息
type CommentUser struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// CommentItem 评论对象（一级评论列表项 / 发表评论响应）
type CommentItem struct {
	ID              string              `json:"id"`
	User            *CommentUser        `json:"user"`
	Content         string              `json:"content"`
	LikeCount       int64               `json:"likeCount"`
	IsLiked         bool                `json:"isLiked"`
	ReplyToID       string              `json:"replyToId"`
	ReplyToNickname string              `json:"replyToNickname"`
	CreatedAt       string              `json:"createdAt"`
	Replies         []*CommentReplyItem `json:"replies"`
	ReplyCount      int64               `json:"replyCount"`
	CanDelete       bool                `json:"canDelete"`
}

// CommentReplyItem 二级评论对象（加载更多返回，无嵌套回复）
type CommentReplyItem struct {
	ID              string       `json:"id"`
	User            *CommentUser `json:"user"`
	Content         string       `json:"content"`
	LikeCount       int64        `json:"likeCount"`
	IsLiked         bool         `json:"isLiked"`
	ReplyToID       string       `json:"replyToId"`
	ReplyToNickname string       `json:"replyToNickname"`
	CreatedAt       string       `json:"createdAt"`
	CanDelete       bool         `json:"canDelete"`
}
