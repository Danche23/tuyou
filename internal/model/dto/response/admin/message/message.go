package message

// MessageItem 后台消息列表项（类型/接收者已映射为中文文案）
type MessageItem struct {
	ID          string `json:"id"`           // 消息id（字符串）
	MsgType     string `json:"msg_type"`     // 消息类型：互动=点赞/关注/评论/收藏；系统=活动/系统通知
	Content     string `json:"content"`      // 消息正文
	ReceiveType string `json:"receive_type"` // 接收类型：全体/指定会员
	Sender      string `json:"sender"`       // 发送者：系统/会员昵称/管理员姓名
	SendTime    string `json:"send_time"`    // 发送时间
}

// ListMessageResp 后台消息分页响应
type ListMessageResp struct {
	List      []MessageItem `json:"list"`
	Total     int64         `json:"total"`
	Page      int           `json:"page"`
	Size      int           `json:"size"`
	TotalPage int           `json:"total_page"`
}
