package search

// SearchRequest 全量搜索请求
type SearchRequest struct {
	Keyword string `form:"keyword" binding:"max=50"`                  // 搜索关键词（必填，Service 层校验非空；最多50字符）
	Type    string `form:"type"`                                      // 搜索类型：content（默认，笔记+攻略）/ activity（活动）
	Page    int    `form:"page"    binding:"omitempty,min=1"`         // 页码（默认 1）
	Size    int    `form:"size"    binding:"omitempty,min=1,max=100"` // 每页条数（默认 10）
}
