package message

// MessageItem 消息列表项
type MessageItem struct {
	ID       string  `json:"id"`       // 消息id（msg_info.id）
	MsgType  int     `json:"msgType"`  // 1=互动消息，2=系统通知
	Content  string  `json:"content"`  // 渲染后的消息正文
	SendTime string  `json:"sendTime"` // 发送时间
	IsRead   int     `json:"isRead"`   // 0未读，1已读
	Sender   *Sender `json:"sender"`   // 发送者信息（系统消息为 null）
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
