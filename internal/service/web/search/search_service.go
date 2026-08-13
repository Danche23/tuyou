package search

import (
	repoSearch "tuyou/internal/repository/web/search"
	"tuyou/pkg/errors"
)

type searchService struct {
	searchRepo repoSearch.SearchRepository
}

// NewSearchService 创建搜索服务实例
func NewSearchService(searchRepo repoSearch.SearchRepository) SearchService {
	return &searchService{searchRepo: searchRepo}
}

// GetHotSearches 获取热门搜索关键词（启用且标记热门的标签名，按 id 升序）
func (s *searchService) GetHotSearches() ([]string, error) {
	names, err := s.searchRepo.FindHotSearchNames()
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询热门搜索关键词失败", err)
	}
	return names, nil
}
