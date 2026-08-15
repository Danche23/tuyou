package message

import (
	"context"
	"time"

	"tuyou/internal/model/entity"
)

// MessageRow 消息列表联查结果行（msg_info LEFT JOIN member/sys_admin 取发送者中文）
type MessageRow struct {
	ID              uint      `gorm:"column:id"`
	MsgType         int       `gorm:"column:msg_type"`
	MsgContent      string    `gorm:"column:msg_content"`
	ReceiveObjType  int       `gorm:"column:receive_obj_type"`
	SenderID        uint      `gorm:"column:sender_id"`
	TemplateID      uint      `gorm:"column:template_id"`
	SendTime        time.Time `gorm:"column:send_time"`
	SenderNickname  string    `gorm:"column:sender_nickname"`
	SenderAdminName string    `gorm:"column:sender_admin_name"`
}

// MessageRepository 后台消息管理仓储接口
type MessageRepository interface {
	// List 分页查询消息主表（按 msg_type + filters 过滤，is_del=0，一条消息一行）
	List(ctx context.Context, msgType int, filters map[string]interface{}, page, size int) ([]MessageRow, int64, error)
	// FindByID 按 id 查询消息主表，不存在返回 gorm.ErrRecordNotFound
	FindByID(ctx context.Context, id uint) (*entity.MsgInfo, error)
	// Create 事务内创建一条消息：msg_info + 批量 msg_user_record（memberIDs 为接收者）
	Create(ctx context.Context, info *entity.MsgInfo, memberIDs []uint) error
	// AllMemberIDs 查询所有会员 id（含领队 role=0 和普通会员 role=1），发全体时扇出接收记录用
	AllMemberIDs(ctx context.Context) ([]uint, error)
	// UnreadCount 查询会员未读消息数（互动/系统拆分），发消息后推角标用
	UnreadCount(ctx context.Context, memberID uint) (interactions, system int64, err error)
	// Delete 硬删除消息主表及其全部接收记录（事务）
	Delete(ctx context.Context, id uint) error
}
