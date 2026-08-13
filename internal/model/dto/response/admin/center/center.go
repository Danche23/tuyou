package center

import "time"

// AdminNotification 管理员个人消息项（后台个人消息，对应 msg_info + msg_user_record）
type AdminNotification struct {
	ID       uint      `json:"id"`
	Type     int       `json:"type"`      // 消息类型（msg_info.msg_type）：2 系统通知
	Content  string    `json:"content"`   // 渲染后的消息正文
	IsRead   int       `json:"is_read"`   // 0未读，1已读
	SendTime time.Time `json:"send_time"` // 消息时间
}

// AdminMsgPage 管理员个人消息分页响应
type AdminMsgPage struct {
	List      []AdminNotification `json:"list"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	Size      int                 `json:"size"`
	TotalPage int                 `json:"total_page"`
}

// AdminUnreadCount 管理员未读消息数响应
type AdminUnreadCount struct {
	UnreadCount int64 `json:"unread_count"`
}
