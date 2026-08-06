package center

import (
	"context"

	response "tuyou/internal/model/dto/response/admin/auth"
)

// CenterService 个人中心服务
type CenterService interface {
	// UpdateAvatar 更新当前管理员头像
	UpdateAvatar(ctx context.Context, adminID uint, avatar string) (*response.AdminInfo, error)

	// UpdateUsername 更新当前管理员账号名
	UpdateUsername(ctx context.Context, adminID uint, username string) (*response.AdminInfo, error)
}
