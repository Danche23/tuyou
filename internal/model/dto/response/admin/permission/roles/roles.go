package roles

// ListRoleResp 角色列表响应
type ListRoleResp struct {
	Total     int            `json:"total"`
	Page      int            `json:"page"`
	Size      int            `json:"size"`
	TotalPage int            `json:"total_page"`
	List      []RoleListItem `json:"list"`
}

// RoleListItem 角色列表项
type RoleListItem struct {
	ID          string `json:"id"`
	RoleName    string `json:"role_name"`
	RoleCode    string `json:"role_code"`
	Description string `json:"description"`
	Status      string `json:"status"` // "启用" 或 "禁用"
	CreateTime  string `json:"create_time"`
}

// InsertRoleResp 新增角色响应
type InsertRoleResp struct {
	ID       string `json:"id"`
	RoleName string `json:"role_name"`
}

// MenuCheckItem 菜单勾选数据（分配树用，仓储层查询结果）
type MenuCheckItem struct {
	ID       uint   `json:"id"`
	MenuName string `json:"menu_name"`
	ParentID uint   `json:"parent_id"`
	Checked  bool   `json:"checked"`
}

// PermCheckItem 接口权限勾选数据（分配树用，仓储层查询结果）
type PermCheckItem struct {
	ID       uint   `json:"id"`
	PermName string `json:"perm_name"`
	Module   string `json:"module"`
	Checked  bool   `json:"checked"`
}

// MenuTreeItem 菜单分配树节点
type MenuTreeItem struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Checked  bool            `json:"checked"`
	Children []*MenuTreeItem `json:"children"`
}

// MenuTreeResp 菜单分配树响应
type MenuTreeResp struct {
	Tree []*MenuTreeItem `json:"menu_tree"`
}

// PermModuleItem 权限分配树 module 分组根节点（无 id，不提交）
type PermModuleItem struct {
	Module   string     `json:"module"` // 所属业务模块名
	Children []PermItem `json:"children"`
}

// PermItem 权限分配树接口权限叶子节点
type PermItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"` // 权限名称
	Checked bool   `json:"checked"`
}

// PermTreeResp 权限分配树响应
type PermTreeResp struct {
	Tree []*PermModuleItem `json:"permission_tree"`
}
