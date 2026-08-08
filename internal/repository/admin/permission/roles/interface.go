package roles

import (
	"context"
	"tuyou/internal/model/dto/response/admin/permission/roles"
	"tuyou/internal/model/entity"
)

type RoleRepository interface {
	// Count 计算总数
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	// List 分页查询角色列表
	List(ctx context.Context, filters map[string]interface{}, page, size int) ([]roles.RoleListItem, error)
	// FindByID 按 ID 查询角色
	FindByID(ctx context.Context, id uint) (*entity.SysRole, error)
	// FindByRoleCode 按角色编码查询（新增/编辑查重用）
	FindByRoleCode(ctx context.Context, roleCode string) (*entity.SysRole, error)
	// Insert 新增角色
	Insert(ctx context.Context, role *entity.SysRole) error
	// Update 更新角色字段
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	// CountAdminsByRoleID 统计引用该角色的管理员数量（删除前校验用）
	CountAdminsByRoleID(ctx context.Context, roleID uint) (int64, error)
	// DeleteCascade 事务内删除角色并清理角色-菜单/权限关联
	DeleteCascade(ctx context.Context, id uint) error

	// ListEnabledMenus 查询全部启用菜单并标记该角色已勾选状态（分配树用）
	ListEnabledMenus(ctx context.Context, roleID uint) ([]roles.MenuCheckItem, error)
	// ListEnabledPerms 查询全部启用接口权限并标记该角色已勾选状态（分配树用）
	ListEnabledPerms(ctx context.Context, roleID uint) ([]roles.PermCheckItem, error)
	// ReplaceRolePerms 事务内覆盖写角色-菜单/权限关联（先删后插）
	ReplaceRolePerms(ctx context.Context, roleID uint, menuIDs, permIDs []uint) error
}
