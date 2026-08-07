package guide

// ListGuideReq 后台攻略分页查询请求
type ListGuideReq struct {
	Page       int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size       int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	Title      string `form:"title"`                                 // 标题模糊搜索
	Category   string `form:"category"`                              // 分类中文名（徒步/露营/骑行/登山/跑步/攀岩/溯溪）
	Difficulty string `form:"difficulty"`                            // 难度（入门/中等/中高等/高等）
	IsSlide    string `form:"is_slide"`                              // 是否轮播：是/否
	SortField  string `form:"sort_field"`                            // 排序字段：热度/浏览量/点赞量/收藏量/评论量/发布时间
	SortOrder  string `form:"sort_order"`                            // 排序方式：升序/降序（默认降序）
	BeginTime  string `form:"begin_time"`                            // 发布开始时间 yyyy-MM-dd
	FinishTime string `form:"finish_time"`                           // 发布结束时间 yyyy-MM-dd
}

// UpdateGuideReq 后台攻略编辑请求（三个字段必传，值未变也要传当前值）
type UpdateGuideReq struct {
	Category string `json:"category" binding:"required"` // 分类中文名（徒步/露营/骑行/登山/跑步/攀岩/溯溪）
	IsSlide  string `json:"is_slide" binding:"required"` // 是否轮播：是/否
	Status   string `json:"status" binding:"required"`   // 上下架：正常/下架
}

// BatchDeleteReq 批量删除攻略请求
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1"` // 攻略 ID 列表（"G0001" 展示格式或纯数字）
}

// ExportGuideReq 批量导出攻略请求
type ExportGuideReq struct {
	IDs []string `json:"ids" binding:"required,min=1"` // 攻略 ID 列表（"G0001" 展示格式或纯数字）
}
