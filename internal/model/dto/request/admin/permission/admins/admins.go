package admins

// ListAdminsReq 查询管理员列表
type ListAdminsReq struct {
	Username string `form:"username"`
	Role     string `form:"role"`
	Status   string `form:"status"` // 空/不传=全部, 启用/禁用 过滤
	Page     uint   `form:"page" binding:"required,min=1"`
	Size     uint   `form:"size" binding:"required,min=1,max=100"`
}

// InsertAdminReq 新增管理员
type InsertAdminReq struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
	Adminname string `json:"adminname" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

// UpdateAdminReq 编辑管理员
type UpdateAdminReq struct {
	Username  string `json:"username"`
	Adminname string `json:"adminname"`
	Role      string `json:"role"`
	Status    string `json:"status"` // 空=不改, 启用/禁用
}

// ResetAdminPasswordReq 重置管理员密码
type ResetAdminPasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
