package activity

// ListActivityReq 后台活动分页查询请求
type ListActivityReq struct {
	Page           int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size           int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	Title          string `form:"title"`                                 // 活动标题模糊搜索
	LeaderName     string `form:"leader_name"`                           // 领队名称模糊搜索
	ActivityType   string `form:"activity_type"`                         // 活动类型中文名（徒步/露营/骑行/登山/跑步/攀岩/溯溪）
	Difficulty     string `form:"difficulty"`                            // 难度（入门/中等/中高等/高等）
	Fee            string `form:"fee"`                                   // 费用类型：免费/付费
	ActivityStatus string `form:"activity_status"`                       // 活动状态：报名中/进行中/已结束
	IsSlide        string `form:"is_slide"`                              // 是否轮播：是/否
	BeginTime      string `form:"begin_time"`                            // 活动开始时间下界 yyyy-MM-dd（start_time >= begin_time）
	FinishTime     string `form:"finish_time"`                           // 活动开始时间上界 yyyy-MM-dd（start_time <= finish_time 23:59:59）
}
