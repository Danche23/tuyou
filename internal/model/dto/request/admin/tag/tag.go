package tag

// ListTagReq 后台标签分页查询请求
type ListTagReq struct {
	Page       int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size       int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	TagName    string `form:"tag_name"`                              // 标签名称模糊搜索
	Status     string `form:"status"`                                // 状态：启用/禁用
	BeginTime  string `form:"begin_time"`                            // 创建时间起始 yyyy-MM-dd
	FinishTime string `form:"finish_time"`                           // 创建时间结束 yyyy-MM-dd
}

// CreateTagReq 后台新增标签请求
type CreateTagReq struct {
	TagName string `json:"tag_name" binding:"required"` // 标签名称
	Status  string `json:"status" binding:"required"`   // 状态：启用/禁用
}

// UpdateTagReq 后台编辑标签请求（status/is_hot 均必填，标签名称不可修改）
type UpdateTagReq struct {
	Status string `json:"status" binding:"required"` // 状态：启用/禁用
	IsHot  string `json:"is_hot" binding:"required"` // 是否热门关键词：是/否
}

// BatchDeleteTagReq 批量删除标签请求
type BatchDeleteTagReq struct {
	IDs []string `json:"ids" binding:"required,min=1"` // 标签 ID 列表（字符串数组）
}
