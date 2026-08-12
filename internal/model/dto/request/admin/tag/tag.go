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
