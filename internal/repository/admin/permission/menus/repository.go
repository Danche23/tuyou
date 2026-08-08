package menus

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"tuyou/internal/model/dto/response/admin/permission/menus"
	"tuyou/internal/model/entity"
	appErrors "tuyou/pkg/errors"
)

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

// applyFilters 拼装动态查询条件
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["menu_name"]; ok {
		query = query.Where("m.menu_name LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["status"]; ok {
		query = query.Where("m.status = ?", v)
	}
	return query
}

// Count 计算总数
func (r *menuRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("sys_menu AS m")
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// List 分页查询菜单列表
func (r *menuRepository) List(ctx context.Context, filters map[string]interface{}, page, size int) ([]menus.MenuListItem, error) {
	var list []menus.MenuListItem

	query := r.db.WithContext(ctx).Table("sys_menu AS m").
		Joins("LEFT JOIN sys_menu AS p ON p.id = m.parent_id").
		Select(`m.id, m.menu_name, m.parent_id, IFNULL(p.menu_name, '') AS parent_name,
			m.menu_type, m.route_path, m.sort,
			IF(m.status = 1, '启用', '禁用') AS status,
			DATE_FORMAT(m.create_time, '%Y-%m-%d') AS create_time`)
	query = applyFilters(query, filters)

	offset := (page - 1) * size
	if err := query.Order("m.sort ASC, m.id ASC").Offset(offset).Limit(size).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListAll 查询全部菜单（父菜单选择器用，不分页）
func (r *menuRepository) ListAll(ctx context.Context) ([]menus.MenuOptionItem, error) {
	var list []menus.MenuOptionItem
	err := r.db.WithContext(ctx).
		Table("sys_menu").
		Select("id, menu_name, parent_id, menu_type, IF(status = 1, '启用', '禁用') AS status").
		Order("sort ASC, id ASC").
		Scan(&list).Error
	return list, err
}

// CountChildren 统计父菜单下的直接子菜单数量
func (r *menuRepository) CountChildren(ctx context.Context, parentID uint) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).Table("sys_menu").Where("parent_id = ?", parentID).Count(&cnt).Error
	return cnt, err
}

// FindByName 按菜单名查询菜单
func (r *menuRepository) FindByName(ctx context.Context, menuName string) (*entity.SysMenu, error) {
	var menu entity.SysMenu
	err := r.db.WithContext(ctx).Where("menu_name = ?", menuName).First(&menu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &menu, nil
}

// FindByID 按 ID 查询菜单
func (r *menuRepository) FindByID(ctx context.Context, id uint) (*entity.SysMenu, error) {
	var menu entity.SysMenu
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&menu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &menu, nil
}

// FindChildrenIDs 批量查询父菜单下的直接子菜单 ID
func (r *menuRepository) FindChildrenIDs(ctx context.Context, parentIDs []uint) ([]uint, error) {
	var ids []uint
	if len(parentIDs) == 0 {
		return ids, nil
	}
	err := r.db.WithContext(ctx).Table("sys_menu").Where("parent_id IN ?", parentIDs).Pluck("id", &ids).Error
	return ids, err
}

// Insert 新增菜单
func (r *menuRepository) Insert(ctx context.Context, menu *entity.SysMenu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

// Update 更新菜单字段
func (r *menuRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Table("sys_menu").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrNotFound
	}
	return nil
}

// DeleteCascade 事务内先删角色-菜单关联，再删菜单
func (r *menuRepository) DeleteCascade(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_menu WHERE menu_id IN ?", ids).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sys_menu WHERE id IN ?", ids).Error; err != nil {
			return err
		}
		return nil
	})
}
