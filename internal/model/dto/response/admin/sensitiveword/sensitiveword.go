package sensitiveword

// Item 敏感词列表项（硬删除语义：删除即物理删行，无启停/软删状态字段）
type Item struct {
	ID         string `json:"id"`          // 词 ID（纯数字字符串）
	Word       string `json:"word"`        // 敏感词内容
	CreateTime string `json:"create_time"` // 创建时间
}

// ListResp 后台敏感词分页响应
type ListResp struct {
	List      []Item `json:"list"`
	Total     int64  `json:"total"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
	TotalPage int    `json:"total_page"`
}
