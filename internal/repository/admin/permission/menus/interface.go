package menus

import (
	"context"
	"tuyou/internal/model/dto/response/admin/permission/menus"
	"tuyou/internal/model/entity"
)

type MenuRepository interface {
	// Count 计算总数
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	// List 分页查询菜单列表
	List(ctx context.Context, filters map[string]interface{}, page, size int) ([]menus.MenuListItem, error)
	// ListAll 查询全部菜单（父菜单选择器用，不分页）
	ListAll(ctx context.Context) ([]menus.MenuOptionItem, error)
	// FindByName 按菜单名查询（重名校验 / 父菜单解析）
	FindByName(ctx context.Context, menuName string) (*entity.SysMenu, error)
	// FindByID 按 ID 查询菜单
	FindByID(ctx context.Context, id uint) (*entity.SysMenu, error)
	// FindChildrenIDs 批量查询父菜单下的直接子菜单 ID
	FindChildrenIDs(ctx context.Context, parentIDs []uint) ([]uint, error)
	// CountChildren 统计父菜单下的直接子菜单数量
	CountChildren(ctx context.Context, parentID uint) (int64, error)
	// Insert 新增菜单
	Insert(ctx context.Context, menu *entity.SysMenu) error
	// Update 更新菜单字段
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	// DeleteCascade 事务内删除菜单及角色-菜单关联
	DeleteCascade(ctx context.Context, ids []uint) error
}
