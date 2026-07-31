package auth

import (
	"context"
	"tuyou/internal/model/entity"
)

type AdminRepository interface {
	// FindByUsername 根据用户名查询用户
	FindByUsername(ctx context.Context, username string) (*entity.SysAdmin, error)

	// FindByID 根据ID查询用户
	FindByID(ctx context.Context, id uint) (*entity.SysAdmin, error)

	// UpdateLastLoginTime 更新最后登录时间
	UpdateLastLoginTime(ctx context.Context, id uint) error

	// FindRoleNameByRoleID 通过角色ID查询角色名称
	FindRoleNameByRoleID(ctx context.Context, roleID uint) (string, error)
}
