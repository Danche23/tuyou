package sensitiveword

// ListReq 后台敏感词分页查询请求
type ListReq struct {
	Page    int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size    int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	Keyword string `form:"keyword"`                               // 词内容模糊搜索，空=不筛
}

// CreateReq 后台新增敏感词请求
type CreateReq struct {
	Word string `json:"word" binding:"required"` // 敏感词内容
}
