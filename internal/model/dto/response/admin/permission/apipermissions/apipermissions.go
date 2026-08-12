package apipermissions

// ListApiPermResp 接口权限列表响应
type ListApiPermResp struct {
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	Size      int               `json:"size"`
	TotalPage int               `json:"total_page"`
	List      []ApiPermListItem `json:"list"`
}

// ApiPermListItem 接口权限列表项
type ApiPermListItem struct {
	ID            string `json:"id"`
	PermName      string `json:"perm_name"`
	Module        string `json:"module"`
	PermKey       string `json:"perm_key"`
	RequestMethod string `json:"request_method"`
	APIPath       string `json:"api_path"`
	Status        string `json:"status"` // "启用" 或 "禁用"
	CreateTime    string `json:"create_time"`
}

// InsertApiPermResp 新增接口权限响应
type InsertApiPermResp struct {
	ID       string `json:"id"`
	PermName string `json:"perm_name"`
}

// ModuleItem 业务模块列表项（获取所有模块下拉用）
type ModuleItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
