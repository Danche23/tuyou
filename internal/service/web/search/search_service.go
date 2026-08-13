package search

import (
	"strings"

	requestSearch "tuyou/internal/model/dto/request/web/search"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
	"tuyou/internal/model/entity"
	repoSearch "tuyou/internal/repository/web/search"
	serviceActivity "tuyou/internal/service/web/activity"
	serviceHome "tuyou/internal/service/web/home"
	"tuyou/pkg/errors"
)

type searchService struct {
	searchRepo  repoSearch.SearchRepository
	homeSvc     serviceHome.HomeService
	activitySvc serviceActivity.ActivityService
}

// NewSearchService 创建搜索服务实例
// 复用 home 的 BuildFeedItems 与 activity 的 BuildActivityItems 组装响应，不重复实现。
func NewSearchService(searchRepo repoSearch.SearchRepository, homeSvc serviceHome.HomeService, activitySvc serviceActivity.ActivityService) SearchService {
	return &searchService{searchRepo: searchRepo, homeSvc: homeSvc, activitySvc: activitySvc}
}

// GetHotSearches 获取热门搜索关键词（启用且标记热门的标签名，按 id 升序）
func (s *searchService) GetHotSearches() ([]string, error) {
	names, err := s.searchRepo.FindHotSearchNames()
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询热门搜索关键词失败", err)
	}
	return names, nil
}

// Search 全量搜索
// type=content（默认）：笔记+攻略，标题/分类/标签 OR 命中，heat DESC 跨表归并 → FeedItem；
// type=activity：活动，标题/分类/地点 OR 命中，start_time DESC → ActivityListItem。
// 不提供 sort/category 参数，排序固定：内容按热度值、活动按开始时间，均倒序。
func (s *searchService) Search(memberID uint, req *requestSearch.SearchRequest) (interface{}, int64, error) {
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return nil, 0, errors.New(errors.CodeInvalidParam, "搜索关键词不能为空")
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	switch req.Type {
	case "", "content":
		return s.searchContent(memberID, keyword, req)
	case "activity":
		return s.searchActivity(memberID, keyword, req)
	default:
		return nil, 0, errors.New(errors.CodeInvalidParam, "type 不合法")
	}
}

// searchContent 内容搜索：笔记+攻略跨表归并
func (s *searchService) searchContent(memberID uint, keyword string, req *requestSearch.SearchRequest) (interface{}, int64, error) {
	offset := (req.Page - 1) * req.Size
	limit := offset + req.Size
	category := categoryStrToInt[keyword] // 0=非分类枚举名，不精确匹配分类

	// 标签匹配：关键词命中启用标签名 → 关联内容ID集合（笔记/攻略各查）
	tagIDs, err := s.searchRepo.FindTagIDsByKeyword(keyword)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}
	noteIDs, err := s.searchRepo.FindNoteIDsByTagIDs(tagIDs)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "查询笔记标签关联失败", err)
	}
	strategyIDs, err := s.searchRepo.FindStrategyIDsByTagIDs(tagIDs)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "查询攻略标签关联失败", err)
	}

	// 各取 offset+size 条（同一排序 heat DESC, id DESC），内存归并后截取当前页
	notes, err := s.searchRepo.SearchNotes(keyword, category, noteIDs, limit)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "搜索笔记失败", err)
	}
	strategies, err := s.searchRepo.SearchStrategies(keyword, category, strategyIDs, limit)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "搜索攻略失败", err)
	}
	noteTotal, err := s.searchRepo.CountSearchNotes(keyword, category, noteIDs)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "统计笔记失败", err)
	}
	strategyTotal, err := s.searchRepo.CountSearchStrategies(keyword, category, strategyIDs)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "统计攻略失败", err)
	}

	merged := mergeContentEntries(notes, strategies)
	start := offset
	if start > len(merged) {
		start = len(merged)
	}
	end := start + req.Size
	if end > len(merged) {
		end = len(merged)
	}

	// 复用 home 组装（唯一实现）：封面/标签/作者/点赞状态/点赞数
	items, err := s.homeSvc.BuildFeedItems(memberID, merged[start:end])
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "组装搜索结果失败", err)
	}
	return items, noteTotal + strategyTotal, nil
}

// searchActivity 活动搜索
func (s *searchService) searchActivity(memberID uint, keyword string, req *requestSearch.SearchRequest) (interface{}, int64, error) {
	offset := (req.Page - 1) * req.Size
	category := categoryStrToInt[keyword] // 0=非分类枚举名，不精确匹配分类

	activities, err := s.searchRepo.SearchActivities(keyword, category, offset, req.Size)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "搜索活动失败", err)
	}
	total, err := s.searchRepo.CountSearchActivities(keyword, category)
	if err != nil {
		return nil, 0, errors.NewWithErr(errors.CodeInternalError, "统计活动失败", err)
	}
	if len(activities) == 0 {
		return []responseActivity.ActivityListItem{}, total, nil
	}

	// 复用 activity 组装（唯一实现）：领队信息/报名状态/分类难度文案
	items, err := s.activitySvc.BuildActivityItems(memberID, activities)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// mergeContentEntries 归并笔记+攻略为精简 FeedEntry
// 两侧已按 heat DESC, id DESC 排序，二路归并得到全局同序序列。
func mergeContentEntries(notes []entity.Note, strategies []entity.Strategy) []serviceHome.FeedEntry {
	merged := make([]serviceHome.FeedEntry, 0, len(notes)+len(strategies))
	i, j := 0, 0
	for i < len(notes) && j < len(strategies) {
		if notes[i].Heat > strategies[j].Heat ||
			(notes[i].Heat == strategies[j].Heat && notes[i].ID >= strategies[j].ID) {
			merged = append(merged, feedEntryFromNote(notes[i]))
			i++
		} else {
			merged = append(merged, feedEntryFromStrategy(strategies[j]))
			j++
		}
	}
	for ; i < len(notes); i++ {
		merged = append(merged, feedEntryFromNote(notes[i]))
	}
	for ; j < len(strategies); j++ {
		merged = append(merged, feedEntryFromStrategy(strategies[j]))
	}
	return merged
}

func feedEntryFromNote(n entity.Note) serviceHome.FeedEntry {
	return serviceHome.FeedEntry{
		IsNote: true, ID: n.ID, MemberID: n.MemberID, Title: n.Title,
		VideoURL: n.VideoURL, Category: n.Category, LikeCount: n.LikeCount,
	}
}

func feedEntryFromStrategy(s entity.Strategy) serviceHome.FeedEntry {
	return serviceHome.FeedEntry{
		IsNote: false, ID: s.ID, MemberID: s.MemberID, Title: s.Title,
		VideoURL: s.VideoURL, Category: s.Category, LikeCount: s.LikeCount,
	}
}

// categoryStrToInt 分类枚举名 → 枚举值（关键词命中分类名时精确匹配）
var categoryStrToInt = map[string]int{
	"徒步": 1, "露营": 2, "骑行": 3, "登山": 4, "跑步": 5, "攀岩": 6, "溯溪": 7,
}
