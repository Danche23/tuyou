package auth

import (
	"context"
	"tuyou/internal/model/entity"
)

type MemberRepository interface {
	// FindByPhone 根据手机号查询用户
	FindByPhone(ctx context.Context, phone string) (*entity.Member, error)

	// FindByID 根据ID查询用户
	FindByID(ctx context.Context, id uint) (*entity.Member, error)

	// Create 创建用户
	Create(ctx context.Context, member *entity.Member) error

	// UpdateLastLoginTime 更新最后登录时间
	UpdateLastLoginTime(ctx context.Context, id uint) error

	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, id uint, password string) error

	// UpdatePhone 更新用户手机号
	UpdatePhone(ctx context.Context, id uint, phone string) error
}
