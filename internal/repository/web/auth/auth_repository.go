package auth

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
	"tuyou/internal/model/entity"
	appErrors "tuyou/pkg/errors"
)

type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &memberRepository{
		db: db,
	}
}

// FindActiveByPhone 查询"在用的手机号"对应的账号（status<>2）
func (r *memberRepository) FindActiveByPhone(ctx context.Context, phone string) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).Where("phone = ? AND status <> 2", phone).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &member, nil
}

// FindCancelledByPhone 查询已注销账号
func (r *memberRepository) FindCancelledByPhone(ctx context.Context, phone string) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).Where("phone = ? AND status = 2", phone).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &member, nil
}

// UpdateStatus 更新账号状态
func (r *memberRepository) UpdateStatus(ctx context.Context, id uint, status int) error {
	return r.db.WithContext(ctx).Model(&entity.Member{}).Where("id = ?", id).Update("status", status).Error
}

// Create 创建用户
func (r *memberRepository) Create(ctx context.Context, member *entity.Member) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// UpdateLastLoginTime 更新用户登录时间
func (r *memberRepository) UpdateLastLoginTime(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Member{}).
		Where("id = ?", id).
		Update("last_login_time", &now).Error
}

// FindByID 根据ID查找用户
func (r *memberRepository) FindByID(ctx context.Context, id uint) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).First(&member, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &member, nil
}

// UpdatePassword 更新用户密码
func (r *memberRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	return r.db.WithContext(ctx).Model(&entity.Member{}).Where("id = ?", id).Update("password", password).Error
}

// UpdatePhone 更新用户手机号
func (r *memberRepository) UpdatePhone(ctx context.Context, id uint, phone string) error {
	return r.db.WithContext(ctx).Model(&entity.Member{}).Where("id = ?", id).Update("phone", phone).Error
}
