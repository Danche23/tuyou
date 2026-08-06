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

// FindMenusByRoleID 查询角色拥有的全部菜单
func (r *adminRepository) FindMenusByRoleID(ctx context.Context, roleID uint) ([]entity.SysMenu, error) {
	var menus []entity.SysMenu
	err := r.db.WithContext(ctx).
		Joins("JOIN sys_role_menu rm ON rm.menu_id = sys_menu.id").
		Where("rm.role_id = ? AND sys_menu.status = 1", roleID).
		Order("sys_menu.sort ASC").
		Find(&menus).Error
	return menus, err
}

// FindPermsByRoleID 查询角色拥有的全部接口权限
func (r *adminRepository) FindPermsByRoleID(ctx context.Context, roleID uint) ([]entity.SysPermission, error) {
	var perms []entity.SysPermission
	err := r.db.WithContext(ctx).
		Joins("JOIN sys_role_perm rp ON rp.perm_id = sys_permission.id").
		Where("rp.role_id = ? AND sys_permission.status = 1", roleID).
		Order("sys_permission.module ASC, sys_permission.id ASC").
		Find(&perms).Error
	return perms, err
}

// UpdatePassword 更新密码
func (r *adminRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	return r.db.WithContext(ctx).
		Model(&entity.SysAdmin{}).
		Where("id = ?", id).
		Update("password", password).Error
}
