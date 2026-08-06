package center

import "context"

// CenterRepository 个人中心数据仓储
type CenterRepository interface {
	// UpdateAvatar 更新头像
	UpdateAvatar(ctx context.Context, id uint, avatar string) error

	// UpdateUsername 更新用户名
	UpdateUsername(ctx context.Context, id uint, username string) error
}
