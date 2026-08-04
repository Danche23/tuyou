package auth

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
	"tuyou/internal/model/entity"
	appErrors "tuyou/pkg/errors"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindByUsername(ctx context.Context, username string) (*entity.SysAdmin, error) {
	var admin entity.SysAdmin
	err := r.db.WithContext(ctx).First(&admin, "username = ?", username).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// FindByID 根据ID查找用户
func (r *adminRepository) FindByID(ctx context.Context, id uint) (*entity.SysAdmin, error) {
	var admin entity.SysAdmin
	err := r.db.WithContext(ctx).First(&admin, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// UpdateLastLoginTime 更新用户登录时间
func (r *adminRepository) UpdateLastLoginTime(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.SysAdmin{}).
		Where("id = ?", id).
		Update("last_login_time", &now).Error
}

// FindRoleNameByRoleID 通过角色ID查询角色名称
func (r *adminRepository) FindRoleNameByRoleID(ctx context.Context, roleID uint) (string, error) {
	var roleName string
	err := r.db.WithContext(ctx).
		Model(&entity.SysRole{}).
		Select("role_name").
		First(&roleName, "id = ?", roleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", appErrors.ErrNotFound
		}
		return "", err
	}
	return roleName, nil
}
