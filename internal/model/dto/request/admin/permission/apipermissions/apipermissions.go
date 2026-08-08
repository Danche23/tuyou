package apipermissions

// ListApiPermsReq 查询接口权限列表
type ListApiPermsReq struct {
	PermName string `form:"perm_name"`
	Module   string `form:"module"`
	Status   string `form:"status"` // 空/不传=全部, 启用/禁用 过滤
	Page     uint   `form:"page" binding:"required,min=1"`
	Size     uint   `form:"size" binding:"required,min=1,max=100"`
}

// InsertApiPermReq 新增接口权限
type InsertApiPermReq struct {
	PermName      string `json:"perm_name" binding:"required"`
	Module        string `json:"module" binding:"required"`
	PermKey       string `json:"perm_key" binding:"required"`
	RequestMethod string `json:"request_method" binding:"required,oneof=GET POST PUT DELETE"`
	APIPath       string `json:"api_path" binding:"required"`
	Status        string `json:"status" binding:"required,oneof=启用 禁用"`
}

// UpdateApiPermReq 编辑接口权限（全部字段必填，整体覆盖）
type UpdateApiPermReq struct {
	PermName      string `json:"perm_name" binding:"required"`
	Module        string `json:"module" binding:"required"`
	PermKey       string `json:"perm_key" binding:"required"`
	RequestMethod string `json:"request_method" binding:"required,oneof=GET POST PUT DELETE"`
	APIPath       string `json:"api_path" binding:"required"`
	Status        string `json:"status" binding:"required,oneof=启用 禁用"`
}
