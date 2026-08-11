package menus

// ListMenuResp 菜单列表响应（一级菜单数组，每项含 children 树）
type ListMenuResp []MenuTreeItem

// MenuTreeItem 菜单树节点
type MenuTreeItem struct {
	MenuListItem
	Children []MenuTreeItem `json:"children"`
}

// MenuListItem 菜单列表项
type MenuListItem struct {
	ID           string `json:"id"`
	MenuName     string `json:"menu_name"`
	ParentID     string `json:"parent_id"`
	ParentName   string `json:"parent_name"`
	MenuType     int    `json:"menu_type"` // 1目录，2菜单
	MenuTypeName string `json:"menu_type_name"`
	RoutePath    string `json:"route_path"`
	Component    string `json:"component"`
	Icon         string `json:"icon"`
	HideInMenu   int    `json:"hide_in_menu"` // 0显示，1隐藏
	Sort         int    `json:"sort"`
	Status       string `json:"status"` // "启用" 或 "禁用"
	CreateTime   string `json:"create_time"`
}

// InsertMenuResp 新增菜单响应
type InsertMenuResp struct {
	ID       string `json:"id"`
	MenuName string `json:"menu_name"`
}

// MenuOptionItem 父菜单选项项（仅一级菜单）
type MenuOptionItem struct {
	ID       string `json:"id"`
	MenuName string `json:"menu_name"`
}
