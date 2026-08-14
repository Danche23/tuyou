package message

import (
	"time"

	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

// MessageRow 消息列表联查结果行（msg_user_record JOIN msg_info JOIN member）
type MessageRow struct {
	ID       uint      `gorm:"column:id"`
	MsgType  int       `gorm:"column:msg_type"`
	Content  string    `gorm:"column:msg_content"`
	SendTime time.Time `gorm:"column:send_time"`
	SenderID uint      `gorm:"column:sender_id"`
	IsRead   int       `gorm:"column:is_read"`
	Nickname string    `gorm:"column:nickname"`
	Avatar   string    `gorm:"column:avatar"`
}

// MessageRepository 消息仓储接口
type MessageRepository interface {
	// FindMessages 查询指定会员某类型消息列表（已读状态来自 msg_user_record）
	FindMessages(memberID uint, msgType, page, size int) ([]MessageRow, int64, error)
	// MarkRead 标记单条消息已读
	MarkRead(memberID, msgID uint) error
	// MarkAllRead 标记全部消息已读
	MarkAllRead(memberID uint) error
	// UnreadCount 未读消息数（按类型分组，返回互动/系统各自未读数）
	UnreadCount(memberID uint) (interactions, system int64, err error)
	// FindCommentIDByMsg 按消息+会员查询关联评论id（校验消息归属，无记录返回 gorm.ErrRecordNotFound）
	FindCommentIDByMsg(msgID, memberID uint) (uint, error)
	// FindCommentByID 查询评论
	FindCommentByID(commentID uint) (*entity.Comment, error)
	// FindMemberNickname 查询会员昵称（消息正文渲染用）
	FindMemberNickname(memberID uint) (string, error)
	// CreateMessage 在事务中插入消息主表
	CreateMessage(tx *gorm.DB, info *entity.MsgInfo) error
	// CreateUserRecord 在事务中插入用户接收记录
	CreateUserRecord(tx *gorm.DB, record *entity.MsgUserRecord) error
	// DeleteUserMessage 软删除消息（is_user_del=1，该会员无此消息记录时返回 gorm.ErrRecordNotFound）
	DeleteUserMessage(memberID, msgID uint) error
}
