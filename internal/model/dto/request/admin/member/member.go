package member

// ListMemberReq 查询会员列表（role 由路由决定：/admin/members→会员，/admin/leaders→领队）
type ListMemberReq struct {
	ID            string `form:"id"`
	Nickname      string `form:"nickname"`
	Phone         string `form:"phone"`
	RegisterStart string `form:"register_start"`
	RegisterEnd   string `form:"register_end"`
	Status        string `form:"status"` // 0/1/2，空=全部
	Page          int    `form:"page" binding:"required,min=1"`
	Size          int    `form:"size" binding:"required,min=1,max=100"`
}

// UpdateMemberReq 编辑会员请求
type UpdateMemberReq struct {
	Status *string `json:"status"` // "启用" 或 "禁用"
	Role   *string `json:"role"`   // "领队" 或 "普通会员"
}

// BatchDeleteReq 批量删除请求
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1,max=500"` // 会员 ID 列表，最多500个
}

// BatchUpdateStatusReq 批量修改状态请求
type BatchUpdateStatusReq struct {
	IDs    []string `json:"ids" binding:"required,min=1"`
	Status string   `json:"status"` // "启用" 或 "禁用"
}
