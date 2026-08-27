package activity

// ListActivityReq 后台活动分页查询请求
type ListActivityReq struct {
	Page           int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size           int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	Title          string `form:"title"`                                 // 活动标题模糊搜索
	LeaderName     string `form:"leader_name"`                           // 领队名称模糊搜索
	ActivityType   string `form:"activity_type"`                         // 活动类型中文名（徒步/露营/骑行/登山/跑步/攀岩/溯溪）
	Difficulty     string `form:"difficulty"`                            // 难度（入门/中等/中高等/高等）
	Fee            string `form:"fee"`                                   // 费用类型：免费/收费
	ActivityStatus string `form:"activity_status"`                       // 活动状态：报名中/进行中/已结束
	IsSlide        string `form:"is_slide"`                              // 是否轮播：是/否
	BeginTime      string `form:"begin_time"`                            // 活动开始时间下界 yyyy-MM-dd（start_time >= begin_time）
	FinishTime     string `form:"finish_time"`                           // 活动开始时间上界 yyyy-MM-dd（start_time <= finish_time 23:59:59）
}

// CreateActivityReq 后台创建活动请求
type CreateActivityReq struct {
	Title           string  `json:"title" binding:"required"`             // 活动标题
	LeaderName      string  `json:"leader_name" binding:"required"`       // 领队名称（对应 /admin/leadernames 返回的 leader_name）
	CoverImg        string  `json:"cover_img"`                            // 活动封面图
	ActivityType    string  `json:"activity_type" binding:"required"`     // 活动类型中文名
	Address         string  `json:"address"`                              // 活动地点
	StartTime       string  `json:"start_time" binding:"required"`        // 活动开始时间 yyyy-MM-dd HH:mm（不含秒）
	EndTime         string  `json:"end_time" binding:"required"`          // 活动结束时间 yyyy-MM-dd HH:mm（不含秒）
	MaxPeople       int     `json:"max_people" binding:"required,min=1"`  // 最大报名人数
	Price           float64 `json:"price"`                                // 费用（元）
	Difficulty      string  `json:"difficulty" binding:"required"`        // 难度中文名
	ActivityContent string  `json:"activity_content" binding:"required"`  // 活动详情内容
	WechatQR        string  `json:"wechat_qr" binding:"required,max=500"` // 微信群二维码图片地址
}

// UpdateActivityReq 后台活动编辑请求（三个字段必传，值未变也要传当前值）
type UpdateActivityReq struct {
	Category string `json:"category" binding:"required"` // 活动类型中文名（徒步/露营/骑行/登山/跑步/攀岩/溯溪）
	IsSlide  string `json:"is_slide" binding:"required"` // 是否轮播：是/否
	Status   string `json:"status" binding:"required"`   // 状态：正常/下架
}

// BatchDeleteReq 批量删除活动请求
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1"` // 活动 ID 列表（"A0001" 展示格式或纯数字）
}

// ExportActivityReq 批量导出活动请求
type ExportActivityReq struct {
	IDs []string `json:"ids" binding:"required,min=1"` // 活动 ID 列表（"A0001" 展示格式或纯数字）
}
