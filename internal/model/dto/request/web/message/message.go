package message

// MessageListRequest 消息列表请求（互动/系统共用）
type MessageListRequest struct {
	Page int `form:"page" json:"page"` // 页码，从1开始，默认1
	Size int `form:"size" json:"size"` // 每页大小，默认10，最大100
}

// ReplyRequest 消息回复评论请求
type ReplyRequest struct {
	Content string `json:"content" binding:"required"` // 回复内容
}
