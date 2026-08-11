package workbench

// WorkbenchResp 后台工作台统计响应
type WorkbenchResp struct {
	MemberTotal   int64               `json:"member_total"`   // 用户总数
	ActivityTotal int64               `json:"activity_total"` // 活动总数
	PublishTotal  int64               `json:"publish_total"`  // 发布总数（笔记+攻略）
	CommentTotal  int64               `json:"comment_total"`  // 评论总数
	RegisterTrend []RegisterTrendItem `json:"register_trend"` // 近7天注册人数（按天）
}

// RegisterTrendItem 近7天注册趋势单日数据
type RegisterTrendItem struct {
	Date  string `json:"date"`  // 日期 YYYY-MM-DD
	Count int64  `json:"count"` // 当日注册人数
}
