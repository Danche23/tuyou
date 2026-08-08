package admins

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"tuyou/internal/model/dto/response/admin/permission/admins"
	"tuyou/internal/model/entity"
	appErrors "tuyou/pkg/errors"
)

type adminsRepository struct {
	db *gorm.DB
}

func NewAdminsRepository(db *gorm.DB) AdminRepository {
	return &adminsRepository{db: db}
}

// 拼装动态查询条件
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["username"]; ok {
		query = query.Where("a.username = ?", v)
	}
	if v, ok := filters["role"]; ok {
		query = query.Where("r.role_name = ?", v)
	}
	if v, ok := filters["status"]; ok {
		query = query.Where("a.status = ?", v)
	}
	return query
}

// Count 计算总数
func (r *adminsRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("sys_admin AS a").
		Joins("LEFT JOIN sys_role AS r ON r.id = a.role_id")
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// List 查询管理员列表
func (r *adminsRepository) List(ctx context.Context, filters map[string]interface{}, page, size int) ([]admins.AdminListItem, error) {
	var list []admins.AdminListItem

	query := r.db.WithContext(ctx).Table("sys_admin AS a").
		Joins("LEFT JOIN sys_role AS r ON r.id = a.role_id").
		Select(`a.id, a.username, IFNULL(a.admin_name, '') AS adminname,
			IFNULL(r.role_name, '') AS role,
			IF(a.status = 1, TRUE, FALSE) AS status,
			DATE_FORMAT(a.create_time, '%Y-%m-%d') AS create_time,
			IFNULL(DATE_FORMAT(a.last_login_time, '%Y-%m-%d %H:%i'), '') AS last_login`)
	query = applyFilters(query, filters)

	offset := (page - 1) * size
	if err := query.Order("a.id DESC").Offset(offset).Limit(size).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByUsername 根据用户名查询管理员，用于新增时查重
func (r *adminsRepository) FindByUsername(ctx context.Context, username string) (*entity.SysAdmin, error) {
	var admin entity.SysAdmin
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &admin, nil
}

// FindRoleIDByRoleName 通过角色名称查询角色ID
func (r *adminsRepository) FindRoleIDByRoleName(ctx context.Context, roleName string) (uint, error) {
	var role entity.SysRole
	err := r.db.WithContext(ctx).Where("role_name = ?", roleName).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, appErrors.ErrNotFound
		}
		return 0, err
	}
	return role.ID, nil
}

// ListRoles 查询所有启用角色
func (r *adminsRepository) ListRoles(ctx context.Context) ([]admins.RoleItem, error) {
	var list []admins.RoleItem
	err := r.db.WithContext(ctx).
		Table("sys_role").
		Select("id, role_name").
		Where("status = 1").
		Order("id ASC").
		Scan(&list).Error
	return list, err
}

// Insert 新增管理员
func (r *adminsRepository) Insert(ctx context.Context, admin *entity.SysAdmin) error {
	return r.db.WithContext(ctx).Create(admin).Error
}

// Update 更新管理员字段
func (r *adminsRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Table("sys_admin").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}

// Delete 删除管理员
func (r *adminsRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Exec("DELETE FROM sys_admin WHERE id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}

// UpdatePassword 重置管理员密码
func (r *adminsRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	return r.db.WithContext(ctx).
		Model(&entity.SysAdmin{}).
		Where("id = ?", id).
		Update("password", password).Error
}
