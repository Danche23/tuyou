package menus

// ListMenusReq 查询菜单列表
type ListMenusReq struct {
	MenuName string `form:"menu_name"`
	Status   string `form:"status"` // 空/不传=全部, 启用/禁用 过滤
	Page     uint   `form:"page" binding:"required,min=1"`
	Size     uint   `form:"size" binding:"required,min=1,max=100"`
}

// InsertMenuReq 新增菜单
type InsertMenuReq struct {
	MenuName   string `json:"menu_name" binding:"required"`
	ParentName string `json:"parent_name"`
	MenuType   string `json:"menu_type" binding:"required,oneof=目录 菜单"`
	RoutePath  string `json:"route_path"`
	Sort       int    `json:"sort"`
	Status     string `json:"status"` // 空=启用
	Component  string `json:"component"`
	Icon       string `json:"icon"`
	HideInMenu bool   `json:"hide_in_menu"` // true=侧边栏隐藏
}

// UpdateMenuReq 编辑菜单（指针字段：不传则不改动）
type UpdateMenuReq struct {
	MenuName   *string `json:"menu_name"`
	ParentName *string `json:"parent_name"` // 空串表示移动到根
	MenuType   *string `json:"menu_type" binding:"omitempty,oneof=目录 菜单"`
	RoutePath  *string `json:"route_path"`
	Sort       *int    `json:"sort"`
	Status     string  `json:"status"` // 空=不改, 启用/禁用
	Component  *string `json:"component"`
	Icon       *string `json:"icon"`
	HideInMenu *bool   `json:"hide_in_menu"` // 不传=不改
}
