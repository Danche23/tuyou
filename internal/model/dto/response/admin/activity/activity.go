package activity

// ActivityListItem 后台活动列表项
type ActivityListItem struct {
	ID             string  `json:"id"`               // 活动展示 ID（"A001"）
	CoverImg       string  `json:"cover_img"`        // 封面图
	Title          string  `json:"title"`            // 活动标题
	LeaderMemberID string  `json:"leader_member_id"` // 领队展示 ID（"M001"）
	LeaderName     string  `json:"leader_name"`      // 领队名称
	ActivityType   string  `json:"activity_type"`    // 活动类型中文名
	StartTime      string  `json:"start_time"`       // 活动开始日期 yyyy-MM-dd
	EndTime        string  `json:"end_time"`         // 活动结束日期 yyyy-MM-dd
	Price          float64 `json:"price"`            // 费用（元）
	CurrentPeople  int     `json:"current_people"`   // 当前已报名人数
	MaxPeople      int     `json:"max_people"`       // 最大容纳报名人数
	Difficulty     string  `json:"difficulty"`       // 难度中文名
	IsSlide        string  `json:"is_slide"`         // 是否轮播：是/否
	ActivityStatus string  `json:"activity_status"`  // 活动状态：报名中/进行中/已结束
	Status         string  `json:"status"`           // 状态：正常/下架
	DeleteTime     string  `json:"delete_time"`      // 删除时间，未删除时为空字符串
}

// ListActivityResp 后台活动分页响应
type ListActivityResp struct {
	List      []ActivityListItem `json:"list"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	Size      int                `json:"size"`
	TotalPage int                `json:"total_page"`
}

// LeaderNameItem 领队名称列表项
type LeaderNameItem struct {
	ID         uint   `json:"id"`          // 领队 ID（原始数字，不带前缀）
	LeaderName string `json:"leader_name"` // 领队名称
}

// ActivityDetailResp 后台活动详情响应（列表项全部字段 + 活动详情内容）
type ActivityDetailResp struct {
	ActivityListItem
	ActivityContent string `json:"activity_content"` // 活动详情内容
}
