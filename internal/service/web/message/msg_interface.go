package message

import (
	requestMsg "tuyou/internal/model/dto/request/web/message"
	responseMsg "tuyou/internal/model/dto/response/web/message"
)

// MessageService 消息服务接口
type MessageService interface {
	// GetInteractions 获取互动消息列表（msg_type=1）
	GetInteractions(memberID uint, req *requestMsg.MessageListRequest) (*responseMsg.MessagePage, error)
	// GetSystem 获取系统通知列表（msg_type=2）
	GetSystem(memberID uint, req *requestMsg.MessageListRequest) (*responseMsg.MessagePage, error)
	// MarkRead 标记单条消息已读
	MarkRead(memberID, msgID uint) error
	// MarkAllRead 标记全部消息已读
	MarkAllRead(memberID uint) error
	// Reply 回复互动消息关联的评论
	Reply(memberID, msgID uint, content string) error
	// UnreadCount 未读消息数（含 total/interactions/system）
	UnreadCount(memberID uint) (*responseMsg.UnreadCount, error)
	// Delete 删除单条消息（软删除）
	Delete(memberID, msgID uint) error
}
