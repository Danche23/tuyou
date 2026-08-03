package member

// ListMemberReq 查询会员列表（role 由路由决定：/admin/members→会员，/admin/leaders→领队）
type ListMemberReq struct {
	ID            string `form:"id"`
	Nickname      string `form:"nickname"`
	Phone         string `form:"phone"`
	RegisterStart string `form:"register_start"`
	RegisterEnd   string `form:"register_end"`
	Page          int    `form:"page" binding:"required,min=1"`
	Size          int    `form:"size" binding:"required,min=1,max=100"`
}

// UpdateMemberReq 编辑会员请求
type UpdateMemberReq struct {
	Status *int `json:"status"` // 0=禁用, 1=正常
	Role   *int `json:"role"`   // 0=领队, 1=普通会员
}
