package search

import (
	requestSearch "tuyou/internal/model/dto/request/web/search"
)

// SearchService 前台全量搜索服务接口
type SearchService interface {
	// GetHotSearches 获取热门搜索关键词（启用且标记热门的标签名，按 id 升序）
	GetHotSearches() ([]string, error)

	// Search 全量搜索：type=content 返回 []FeedItem（笔记+攻略跨表合并，发布时间倒序）；type=activity 返回 []ActivityListItem
	// 返回 (list, total, err)；controller 用 response.NewPageResponse 包装分页
	Search(memberID uint, req *requestSearch.SearchRequest) (interface{}, int64, error)
}
