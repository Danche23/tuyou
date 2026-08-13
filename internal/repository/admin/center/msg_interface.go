package center

import (
	"context"
	"time"
)

// AdminMsgRow 后台个人消息联查结果行（msg_user_record JOIN msg_info，receive_obj_type=2）
type AdminMsgRow struct {
	ID       uint      `gorm:"column:id"`
	Type     int       `gorm:"column:type"`
	Content  string    `gorm:"column:content"`
	IsRead   int       `gorm:"column:is_read"`
	SendTime time.Time `gorm:"column:send_time"`
}

// AdminMsgRepository 后台管理员个人消息仓储
// 复用 msg_info + msg_user_record，通过 msg_info.receive_obj_type=2 区分管理员个人消息
type AdminMsgRepository interface {
	// Create 生成一条管理员系统通知（msg_info + msg_user_record 同一事务），返回消息ID
	Create(ctx context.Context, adminID uint, content string) error
	// List 分页查询管理员个人消息
	List(ctx context.Context, adminID uint, page, size int) ([]AdminMsgRow, int64, error)
	// UnreadCount 未读消息数
	UnreadCount(ctx context.Context, adminID uint) (int64, error)
	// MarkRead 标记单条消息已读；无该消息记录返回 gorm.ErrRecordNotFound
	MarkRead(ctx context.Context, adminID, msgID uint) error
	// MarkAllRead 标记该管理员全部消息已读
	MarkAllRead(ctx context.Context, adminID uint) error
	// Delete 软删除单条消息（is_user_del=1）；无该消息记录返回 gorm.ErrRecordNotFound
	Delete(ctx context.Context, adminID, msgID uint) error
	// ClearAll 清空该管理员全部消息（软删除）
	ClearAll(ctx context.Context, adminID uint) error
}
