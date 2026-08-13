package search

// SearchService 前台全量搜索服务接口
type SearchService interface {
	// GetHotSearches 获取热门搜索关键词（启用且标记热门的标签名，按 id 升序）
	GetHotSearches() ([]string, error)
}
