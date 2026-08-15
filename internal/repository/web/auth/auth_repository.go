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

// FindByPhone 根据手机号查询用户
func (r *memberRepository) FindByPhone(ctx context.Context, phone string) (*entity.Member, error) {
	var member entity.Member
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &member, nil
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
