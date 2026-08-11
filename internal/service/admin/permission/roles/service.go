package roles

import (
	"context"
	"errors"
	"math"
	"strconv"

	request "tuyou/internal/model/dto/request/admin/permission/roles"
	response "tuyou/internal/model/dto/response/admin/permission/roles"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository/admin/permission/roles"
	appErrors "tuyou/pkg/errors"
)

type roleService struct {
	roleRepo roles.RoleRepository
}

func NewRoleService(roleRepo roles.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

// statusToInt 状态中文转 int：启用=1，禁用=0
func statusToInt(s string) int {
	if s == "禁用" {
		return 0
	}
	return 1
}

// strToUintSlice 字符串 id 数组转 uint 数组
func strToUintSlice(ids []string) ([]uint, error) {
	result := make([]uint, 0, len(ids))
	for _, s := range ids {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, appErrors.ErrInvalidParam
		}
		result = append(result, uint(v))
	}
	return result, nil
}

// ListRoles 分页查询角色列表
func (s *roleService) ListRoles(ctx context.Context, req *request.ListRolesReq) (*response.ListRoleResp, error) {
	filters := map[string]interface{}{}
	if req.RoleName != "" {
		filters["role_name"] = req.RoleName
	}
	if req.Status != "" {
		filters["status"] = statusToInt(req.Status)
	}

	total, err := s.roleRepo.Count(ctx, filters)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	list, err := s.roleRepo.List(ctx, filters, int(req.Page), int(req.Size))
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	return &response.ListRoleResp{
		Total:     int(total),
		Page:      int(req.Page),
		Size:      int(req.Size),
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
		List:      list,
	}, nil
}

// InsertRole 新增角色
func (s *roleService) InsertRole(ctx context.Context, req *request.InsertRoleReq) (*response.InsertRoleResp, error) {
	// role_code 查重：DB 真错误原样返回，不存在则继续
	exist, err := s.roleRepo.FindByRoleCode(ctx, req.RoleCode)
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, appErrors.New(appErrors.CodeResourceAlreadyExists, "角色编码已存在")
	}

	role := &entity.SysRole{
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
		Status:      1,
	}
	if err := s.roleRepo.Insert(ctx, role); err != nil {
		return nil, err
	}
	return &response.InsertRoleResp{
		ID:       strconv.Itoa(int(role.ID)),
		RoleName: role.RoleName,
	}, nil
}

// UpdateRole 编辑角色
func (s *roleService) UpdateRole(ctx context.Context, id uint, req *request.UpdateRoleReq) error {
	// 校验角色存在
	if _, err := s.roleRepo.FindByID(ctx, id); err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if req.RoleName != "" {
		updates["role_name"] = req.RoleName
	}
	if req.RoleCode != "" {
		// role_code 改动时查重，排除自身
		exist, err := s.roleRepo.FindByRoleCode(ctx, req.RoleCode)
		if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
			return err
		}
		if exist != nil && exist.ID != id {
			return appErrors.New(appErrors.CodeResourceAlreadyExists, "角色编码已存在")
		}
		updates["role_code"] = req.RoleCode
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Status != "" {
		updates["status"] = statusToInt(req.Status)
	}

	if len(updates) == 0 {
		return appErrors.ErrMissingParam
	}

	return s.roleRepo.Update(ctx, id, updates)
}

// DeleteRole 删除角色（级联清理菜单/权限关联）
func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	// 校验角色存在
	if _, err := s.roleRepo.FindByID(ctx, id); err != nil {
		return err
	}
	// 存在引用该角色的管理员时禁止删除，避免管理员权限悬空
	cnt, err := s.roleRepo.CountAdminsByRoleID(ctx, id)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return appErrors.New(appErrors.CodeResourceLocked, "该角色下存在管理员，无法删除")
	}
	return s.roleRepo.DeleteCascade(ctx, id)
}

// GetMenuTree 查询角色菜单分配树
func (s *roleService) GetMenuTree(ctx context.Context, roleID uint) (*response.MenuTreeResp, error) {
	if _, err := s.roleRepo.FindByID(ctx, roleID); err != nil {
		return nil, err
	}
	items, err := s.roleRepo.ListEnabledMenus(ctx, roleID)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}
	return &response.MenuTreeResp{Tree: buildMenuTree(items)}, nil
}

// GetPermTree 查询角色权限分配树
func (s *roleService) GetPermTree(ctx context.Context, roleID uint) (*response.PermTreeResp, error) {
	if _, err := s.roleRepo.FindByID(ctx, roleID); err != nil {
		return nil, err
	}
	items, err := s.roleRepo.ListEnabledPerms(ctx, roleID)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}
	return &response.PermTreeResp{Tree: buildPermTree(items)}, nil
}

// AssignPermissions 角色分配菜单与接口权限（全量覆盖）
func (s *roleService) AssignPermissions(ctx context.Context, roleID uint, req *request.AssignPermissionsReq) error {
	if _, err := s.roleRepo.FindByID(ctx, roleID); err != nil {
		return err
	}
	menuIDs, err := strToUintSlice(req.MenuIDs)
	if err != nil {
		return err
	}
	permIDs, err := strToUintSlice(req.PermIDs)
	if err != nil {
		return err
	}
	return s.roleRepo.ReplaceRolePerms(ctx, roleID, menuIDs, permIDs)
}

// buildMenuTree 将带勾选状态的扁平菜单列表组装成树
func buildMenuTree(items []response.MenuCheckItem) []*response.MenuTreeItem {
	nodes := make(map[uint]*response.MenuTreeItem, len(items))
	roots := make([]*response.MenuTreeItem, 0, len(items))
	for i := range items {
		it := &items[i]
		nodes[it.ID] = &response.MenuTreeItem{
			ID:       strconv.Itoa(int(it.ID)),
			Name:     it.MenuName,
			Checked:  it.Checked,
			Children: []*response.MenuTreeItem{},
		}
	}
	for _, it := range items {
		node := nodes[it.ID]
		if it.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodes[it.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 父菜单不存在（如父被禁用）时挂到根，避免节点丢失
			roots = append(roots, node)
		}
	}
	return roots
}

// buildPermTree 将带勾选状态的接口权限按 module 分组组装成树
func buildPermTree(items []response.PermCheckItem) []*response.PermModuleItem {
	var roots []*response.PermModuleItem
	idx := make(map[string]int)
	for _, it := range items {
		i, ok := idx[it.Module]
		if !ok {
			roots = append(roots, &response.PermModuleItem{
				Module:   it.Module,
				Children: []response.PermItem{},
			})
			i = len(roots) - 1
			idx[it.Module] = i
		}
		roots[i].Children = append(roots[i].Children, response.PermItem{
			ID:      strconv.Itoa(int(it.ID)),
			Name:    it.PermName,
			Checked: it.Checked,
		})
	}
	return roots
}
