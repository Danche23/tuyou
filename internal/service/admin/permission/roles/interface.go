package roles

import (
	"context"
	request "tuyou/internal/model/dto/request/admin/permission/roles"
	response "tuyou/internal/model/dto/response/admin/permission/roles"
)

type RoleService interface {
	// ListRoles 分页查询角色列表
	ListRoles(ctx context.Context, req *request.ListRolesReq) (*response.ListRoleResp, error)
	// InsertRole 新增角色
	InsertRole(ctx context.Context, req *request.InsertRoleReq) (*response.InsertRoleResp, error)
	// UpdateRole 编辑角色
	UpdateRole(ctx context.Context, id uint, req *request.UpdateRoleReq) error
	// DeleteRole 删除角色（级联清理菜单/权限关联）
	DeleteRole(ctx context.Context, id uint) error
	// GetMenuTree 查询角色菜单分配树（全量启用菜单 + 勾选状态）
	GetMenuTree(ctx context.Context, roleID uint) (*response.MenuTreeResp, error)
	// GetPermTree 查询角色权限分配树（按 module 分组 + 勾选状态）
	GetPermTree(ctx context.Context, roleID uint) (*response.PermTreeResp, error)
	// AssignPermissions 角色分配菜单与接口权限（全量覆盖）
	AssignPermissions(ctx context.Context, roleID uint, req *request.AssignPermissionsReq) error
}
