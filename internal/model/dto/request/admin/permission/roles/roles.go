package roles

// ListRolesReq 查询角色列表
type ListRolesReq struct {
	RoleName string `form:"role_name"`
	Status   string `form:"status"` // 空/不传=全部, 启用/禁用 过滤
	Page     uint   `form:"page" binding:"required,min=1"`
	Size     uint   `form:"size" binding:"required,min=1,max=100"`
}

// InsertRoleReq 新增角色
type InsertRoleReq struct {
	RoleName    string `json:"role_name" binding:"required"`
	RoleCode    string `json:"role_code" binding:"required"`
	Description string `json:"description"`
}

// UpdateRoleReq 编辑角色（空字段不修改）
type UpdateRoleReq struct {
	RoleName    string `json:"role_name"`
	RoleCode    string `json:"role_code"`
	Description string `json:"description"`
	Status      string `json:"status"` // 空=不改, 启用/禁用
}

// AssignPermissionsReq 角色权限分配（全量覆盖）
type AssignPermissionsReq struct {
	MenuIDs []string `json:"menu_ids"`
	PermIDs []string `json:"perm_ids"`
}
