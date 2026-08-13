package tag

// TagListItem 后台标签列表项
type TagListItem struct {
	ID         string `json:"id"`          // 标签 ID（字符串）
	TagName    string `json:"tag_name"`    // 标签名称
	Status     string `json:"status"`      // 状态：启用/禁用
	IsHot      string `json:"is_hot"`      // 是否热门关键词：是/否
	CreatedAt  string `json:"created_at"`  // 创建时间
	DeleteTime string `json:"delete_time"` // 删除时间（禁用/软删时非空），未删除时为空字符串
}

// ListTagResp 后台标签分页响应
type ListTagResp struct {
	List      []TagListItem `json:"list"`
	Total     int64         `json:"total"`
	Page      int           `json:"page"`
	Size      int           `json:"size"`
	TotalPage int           `json:"total_page"`
}
