package roles

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"tuyou/internal/model/dto/response/admin/permission/roles"
	"tuyou/internal/model/entity"
	appErrors "tuyou/pkg/errors"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// applyFilters 拼装动态查询条件
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["role_name"]; ok {
		query = query.Where("role_name LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["status"]; ok {
		query = query.Where("status = ?", v)
	}
	return query
}

// Count 计算总数
func (r *roleRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("sys_role")
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// List 分页查询角色列表
func (r *roleRepository) List(ctx context.Context, filters map[string]interface{}, page, size int) ([]roles.RoleListItem, error) {
	var list []roles.RoleListItem
	query := r.db.WithContext(ctx).Table("sys_role").
		Select(`id, role_name, role_code, IFNULL(description, '') AS description,
			IF(status = 1, '启用', '禁用') AS status,
			DATE_FORMAT(create_time, '%Y-%m-%d') AS create_time`)
	query = applyFilters(query, filters)

	offset := (page - 1) * size
	if err := query.Order("id DESC").Offset(offset).Limit(size).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByID 按 ID 查询角色
func (r *roleRepository) FindByID(ctx context.Context, id uint) (*entity.SysRole, error) {
	var role entity.SysRole
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// FindByRoleCode 按角色编码查询（新增/编辑查重用）
func (r *roleRepository) FindByRoleCode(ctx context.Context, roleCode string) (*entity.SysRole, error) {
	var role entity.SysRole
	err := r.db.WithContext(ctx).Where("role_code = ?", roleCode).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// Insert 新增角色
func (r *roleRepository) Insert(ctx context.Context, role *entity.SysRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Update 更新角色字段
func (r *roleRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Table("sys_role").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrNotFound
	}
	return nil
}

// CountAdminsByRoleID 统计引用该角色的管理员数量
func (r *roleRepository) CountAdminsByRoleID(ctx context.Context, roleID uint) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Table("sys_admin").
		Where("role_id = ?", roleID).
		Count(&cnt).Error
	return cnt, err
}

// DeleteCascade 事务内删除角色并清理角色-菜单/权限关联
func (r *roleRepository) DeleteCascade(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_menu WHERE role_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sys_role_perm WHERE role_id = ?", id).Error; err != nil {
			return err
		}
		result := tx.Exec("DELETE FROM sys_role WHERE id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return appErrors.ErrNotFound
		}
		return nil
	})
}

// ListEnabledMenus 查询全部启用菜单并标记该角色已勾选状态（分配树用）
func (r *roleRepository) ListEnabledMenus(ctx context.Context, roleID uint) ([]roles.MenuCheckItem, error) {
	var list []roles.MenuCheckItem
	err := r.db.WithContext(ctx).Table("sys_menu AS m").
		Joins("LEFT JOIN sys_role_menu AS rm ON rm.menu_id = m.id AND rm.role_id = ?", roleID).
		Select(`m.id, m.menu_name, m.parent_id, IF(rm.role_id IS NOT NULL, 1, 0) AS checked`).
		Where("m.status = 1").
		Order("m.sort ASC, m.id ASC").
		Scan(&list).Error
	return list, err
}

// ListEnabledPerms 查询全部启用接口权限并标记该角色已勾选状态（分配树用）
func (r *roleRepository) ListEnabledPerms(ctx context.Context, roleID uint) ([]roles.PermCheckItem, error) {
	var list []roles.PermCheckItem
	err := r.db.WithContext(ctx).Table("sys_permission AS p").
		Joins("LEFT JOIN sys_role_perm AS rp ON rp.perm_id = p.id AND rp.role_id = ?", roleID).
		Select(`p.id, p.perm_name, p.module, IF(rp.role_id IS NOT NULL, 1, 0) AS checked`).
		Where("p.status = 1").
		Order("p.module ASC, p.id ASC").
		Scan(&list).Error
	return list, err
}

// ReplaceRolePerms 事务内覆盖写角色-菜单/权限关联（先删后插）
func (r *roleRepository) ReplaceRolePerms(ctx context.Context, roleID uint, menuIDs, permIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_menu WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sys_role_perm WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}
		if len(menuIDs) > 0 {
			rows := make([]entity.SysRoleMenu, 0, len(menuIDs))
			for _, mid := range menuIDs {
				rows = append(rows, entity.SysRoleMenu{RoleID: roleID, MenuID: mid})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
		if len(permIDs) > 0 {
			rows := make([]entity.SysRolePerm, 0, len(permIDs))
			for _, pid := range permIDs {
				rows = append(rows, entity.SysRolePerm{RoleID: roleID, PermID: pid})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
