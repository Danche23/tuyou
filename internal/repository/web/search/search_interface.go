package search

// SearchRepository 前台全量搜索仓储接口
type SearchRepository interface {
	// FindHotSearchNames 查询热门搜索关键词（status=1 AND is_hot=1 且未删除），按 id 升序
	FindHotSearchNames() ([]string, error)
}
