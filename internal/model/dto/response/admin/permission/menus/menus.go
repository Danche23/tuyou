package menus

// ListMenuResp 菜单列表响应
type ListMenuResp struct {
	Total     int            `json:"total"`
	Page      int            `json:"page"`
	Size      int            `json:"size"`
	TotalPage int            `json:"total_page"`
	List      []MenuListItem `json:"list"`
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
	Sort         int    `json:"sort"`
	Status       string `json:"status"` // "启用" 或 "禁用"
	CreateTime   string `json:"create_time"`
}

// InsertMenuResp 新增菜单响应
type InsertMenuResp struct {
	ID       string `json:"id"`
	MenuName string `json:"menu_name"`
}

// MenuOptionItem 全部菜单选项项（父菜单选择器用）
type MenuOptionItem struct {
	ID       string `json:"id"`
	MenuName string `json:"menu_name"`
	ParentID string `json:"parent_id"`
	MenuType int    `json:"menu_type"` // 1目录，2菜单
	Status   string `json:"status"`    // "启用" 或 "禁用"
}
