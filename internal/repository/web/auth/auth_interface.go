package auth

import (
	"context"
	"tuyou/internal/model/entity"
)

type MemberRepository interface {
	// FindActiveByPhone 查询"在用的手机号"（status<>2）对应的账号（注销后手机号可被重新注册）
	FindActiveByPhone(ctx context.Context, phone string) (*entity.Member, error)

	// FindCancelledByPhone 查询已注销（status=2）账号（用于区分"未注册"与"已注销"）
	FindCancelledByPhone(ctx context.Context, phone string) (*entity.Member, error)

	// UpdateStatus 更新账号状态（注销=2 / 恢复=1 / 禁用=0）
	UpdateStatus(ctx context.Context, id uint, status int) error

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
