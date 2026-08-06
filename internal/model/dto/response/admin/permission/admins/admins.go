package admins

// ListAdminResp 管理员列表响应
type ListAdminResp struct {
	Total     int             `json:"total"`
	Page      int             `json:"page"`
	Size      int             `json:"size"`
	TotalPage int             `json:"total_page"`
	List      []AdminListItem `json:"list"`
}

// AdminListItem 管理员列表项
type AdminListItem struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Adminname  string `json:"adminname"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	CreateTime string `json:"create_time"`
	LastLogin  string `json:"last_login"`
}

// InsertAdminResp 增加管理员响应
type InsertAdminResp struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

// RoleItem 角色列表项
type RoleItem struct {
	ID       uint   `json:"id"`
	RoleName string `json:"role_name"`
}
