package center

import (
	"context"

	response "tuyou/internal/model/dto/response/admin/auth"
	responseCenter "tuyou/internal/model/dto/response/admin/center"
)

// CenterService 个人中心服务
type CenterService interface {
	// UpdateAvatar 更新当前管理员头像
	UpdateAvatar(ctx context.Context, adminID uint, avatar string) (*response.AdminInfo, error)

	// UpdateUsername 更新当前管理员账号名
	UpdateUsername(ctx context.Context, adminID uint, username string) (*response.AdminInfo, error)

	// GetNotifications 分页获取当前管理员个人消息
	GetNotifications(ctx context.Context, adminID uint, page, size int) (*responseCenter.AdminMsgPage, error)
	// GetUnreadCount 获取当前管理员未读消息数
	GetUnreadCount(ctx context.Context, adminID uint) (*responseCenter.AdminUnreadCount, error)
	// MarkRead 标记单条消息已读
	MarkRead(ctx context.Context, adminID, msgID uint) error
	// MarkAllRead 标记全部消息已读
	MarkAllRead(ctx context.Context, adminID uint) error
	// Delete 删除单条消息
	Delete(ctx context.Context, adminID, msgID uint) error
	// ClearAll 清空所有消息
	ClearAll(ctx context.Context, adminID uint) error
}
