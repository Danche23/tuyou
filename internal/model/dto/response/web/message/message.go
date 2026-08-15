package message

// MessageItem 消息列表项
type MessageItem struct {
	ID       string          `json:"id"`       // 消息id（msg_info.id）
	MsgType  int             `json:"msgType"`  // 1=互动消息，2=系统通知
	Content  string          `json:"content"`  // 渲染后的消息正文
	SendTime string          `json:"sendTime"` // 发送时间
	IsRead   int             `json:"isRead"`   // 0未读，1已读
	Sender   *Sender         `json:"sender"`   // 发送者信息（系统消息为 null）
	Comment  *CommentMessage `json:"comment"`  // 评论类消息（comment_id>0）才有，展开用；非评论类为 null
}

// CommentMessage 评论类消息展开后展示的评论内容。
// 顶层 sender="you"（本条消息=发信人发的）；replies 只含"我"发的回复（sender="me"），时间正序。
type CommentMessage struct {
	CommentID string            `json:"commentId"` // 评论id（string，防 JS 精度丢失）
	Sender    string            `json:"sender"`    // "you"=对方发的消息，"me"=我发的回复
	Content   string            `json:"content"`   // 评论原始内容
	SendTime  string            `json:"sendTime"`  // 创建时间 yyyy-MM-dd HH:mm:ss
	LikeCount int64             `json:"likeCount"` // 点赞数
	IsLiked   bool              `json:"isLiked"`   // 当前登录用户是否已点赞
	Replies   []*CommentMessage `json:"replies"`   // 回复列表：只含"我"发的回复
}

// Sender 发送者信息
type Sender struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// MessagePage 消息分页响应
type MessagePage struct {
	List      []MessageItem `json:"list"`
	Total     int64         `json:"total"`
	Page      int           `json:"page"`
	Size      int           `json:"size"`
	TotalPage int           `json:"total_page"`
}

// UnreadCount 未读消息数响应
type UnreadCount struct {
	Total        int64 `json:"total"`        // 未读消息总数
	Interactions int64 `json:"interactions"` // 互动消息未读数
	System       int64 `json:"system"`       // 系统通知未读数
}
