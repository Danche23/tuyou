package home

import (
	"context"
	"math"
	"sort"
	"strconv"
	"time"

	requestHome "tuyou/internal/model/dto/request/web/home"
	responseHome "tuyou/internal/model/dto/response/web/home"
	"tuyou/internal/model/entity"
	repoGuide "tuyou/internal/repository/web/guide"
	repoHome "tuyou/internal/repository/web/home"
	repoNote "tuyou/internal/repository/web/note"
	"tuyou/pkg/errors"
	"tuyou/pkg/sensitive"
)

type homeService struct {
	homeRepo  repoHome.HomeRepository
	noteRepo  repoNote.NoteRepository
	guideRepo repoGuide.GuideRepository
	sensitive *sensitive.SensitiveMatcher
}

// NewHomeService 创建首页服务实例
// 注入 note/guide 仓储以复用点赞状态缓存（Redis 优先）
func NewHomeService(homeRepo repoHome.HomeRepository, noteRepo repoNote.NoteRepository, guideRepo repoGuide.GuideRepository, matcher *sensitive.SensitiveMatcher) HomeService {
	return &homeService{homeRepo: homeRepo, noteRepo: noteRepo, guideRepo: guideRepo, sensitive: matcher}
}

// GetSlides 获取首页轮播（固定 5 张）
//
// 填充顺序：
//
//	① is_slide=1 的活动 → ② is_slide=1 的笔记 → ③ is_slide=1 的攻略
//	④ 不足 5 张时从活动表补齐（is_slide=0）→ ⑤ 从笔记表补齐 → ⑥ 从攻略表补齐
//
// 补齐为显示时自动选中、不写库；排除已删除/已结束活动、下架笔记/攻略；内容不足 5 张时有多少返多少。
func (s *homeService) GetSlides() ([]responseHome.SlideItem, error) {
	// ========== ①②③ 管理员已标记的轮播 ==========
	slideActs, err := s.homeRepo.FindSlideActivities()
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询轮播活动失败", err)
	}
	slideNotes, err := s.homeRepo.FindSlideNotes()
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询轮播笔记失败", err)
	}
	slideStrategies, err := s.homeRepo.FindSlideStrategies()
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询轮播攻略失败", err)
	}

	count := len(slideActs) + len(slideNotes) + len(slideStrategies)

	// ========== ④⑤⑥ 不足 5 张时补齐（不写库） ==========
	var fillActs []entity.Activity
	var fillNotes []entity.Note
	var fillStrategies []entity.Strategy
	if count < 5 {
		remaining := 5 - count
		fillActs, err = s.homeRepo.FindFillActivities(remaining)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "补齐轮播活动失败", err)
		}
		remaining -= len(fillActs)
		if remaining > 0 {
			fillNotes, err = s.homeRepo.FindFillNotes(remaining)
			if err != nil {
				return nil, errors.NewWithErr(errors.CodeInternalError, "补齐轮播笔记失败", err)
			}
			remaining -= len(fillNotes)
		}
		if remaining > 0 {
			fillStrategies, err = s.homeRepo.FindFillStrategies(remaining)
			if err != nil {
				return nil, errors.NewWithErr(errors.CodeInternalError, "补齐轮播攻略失败", err)
			}
		}
	}

	// ========== 批量取笔记/攻略的封面与标签（活动字段在主表，无需关联查询） ==========
	noteIDs := make([]uint, 0, len(slideNotes)+len(fillNotes))
	noteIDs = append(noteIDs, collectNoteIDs(slideNotes)...)
	noteIDs = append(noteIDs, collectNoteIDs(fillNotes)...)

	strategyIDs := make([]uint, 0, len(slideStrategies)+len(fillStrategies))
	strategyIDs = append(strategyIDs, collectStrategyIDs(slideStrategies)...)
	strategyIDs = append(strategyIDs, collectStrategyIDs(fillStrategies)...)

	noteCovers, err := s.homeRepo.FindFirstImagesByNoteIDs(noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记封面失败", err)
	}
	strategyCovers, err := s.homeRepo.FindFirstImagesByStrategyIDs(strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略封面失败", err)
	}
	noteTags, err := s.homeRepo.FindFirstTagsByNoteIDs(noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记标签失败", err)
	}
	strategyTags, err := s.homeRepo.FindFirstTagsByStrategyIDs(strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略标签失败", err)
	}

	// ========== 组装（按优先级顺序） ==========
	items := make([]responseHome.SlideItem, 0, 5)
	for _, act := range slideActs {
		items = append(items, s.toActivitySlide(act))
	}
	for _, n := range slideNotes {
		items = append(items, s.toNoteSlide(n, noteCovers[n.ID], noteTags[n.ID]))
	}
	for _, st := range slideStrategies {
		items = append(items, s.toStrategySlide(st, strategyCovers[st.ID], strategyTags[st.ID]))
	}
	for _, act := range fillActs {
		items = append(items, s.toActivitySlide(act))
	}
	for _, n := range fillNotes {
		items = append(items, s.toNoteSlide(n, noteCovers[n.ID], noteTags[n.ID]))
	}
	for _, st := range fillStrategies {
		items = append(items, s.toStrategySlide(st, strategyCovers[st.ID], strategyTags[st.ID]))
	}

	// 防御性截断，最多 5 张
	if len(items) > 5 {
		items = items[:5]
	}
	return items, nil
}

// toActivitySlide 活动 → 轮播项（字段全部来自活动主表）
func (s *homeService) toActivitySlide(act entity.Activity) responseHome.SlideItem {
	return responseHome.SlideItem{
		ID:            strconv.FormatUint(uint64(act.ID), 10),
		Type:          "activity",
		Title:         s.mask(act.Title),
		Cover:         act.CoverImg,
		Tag:           activityCategory[act.ActivityType],
		SignUpStart:   act.SignUpStart.Format("2006-01-02 15:04:05"),
		CurrentPeople: act.CurrentPeople,
		MaxPeople:     act.MaxPeople,
	}
}

// toNoteSlide 笔记 → 轮播项（封面/标签来自批量查询结果）
// 降级标签渲染：无启用标签时兜底为内容分类名
func (s *homeService) toNoteSlide(n entity.Note, cover, tag string) responseHome.SlideItem {
	if tag == "" {
		tag = activityCategory[n.Category]
	}
	return responseHome.SlideItem{
		ID:    strconv.FormatUint(uint64(n.ID), 10),
		Type:  "note",
		Title: s.mask(n.Title),
		Cover: cover,
		Tag:   tag,
	}
}

// toStrategySlide 攻略 → 轮播项（封面/标签来自批量查询结果）
// 降级标签渲染：无启用标签时兜底为内容分类名
func (s *homeService) toStrategySlide(st entity.Strategy, cover, tag string) responseHome.SlideItem {
	if tag == "" {
		tag = activityCategory[st.Category]
	}
	return responseHome.SlideItem{
		ID:    strconv.FormatUint(uint64(st.ID), 10),
		Type:  "guide",
		Title: s.mask(st.Title),
		Cover: cover,
		Tag:   tag,
	}
}

func collectNoteIDs(notes []entity.Note) []uint {
	ids := make([]uint, 0, len(notes))
	for _, n := range notes {
		ids = append(ids, n.ID)
	}
	return ids
}

func collectStrategyIDs(strategies []entity.Strategy) []uint {
	ids := make([]uint, 0, len(strategies))
	for _, st := range strategies {
		ids = append(ids, st.ID)
	}
	return ids
}

// GetFeed 获取Feed流（笔记/攻略跨表合并，瀑布流分页）
//
// 排序：推荐=热门标签内容置顶→发布时间倒序（无热门标签时退化为纯发布时间倒序）；最新=发布时间倒序；最热=热度倒序→点赞数倒序。
// 分页：笔记/攻略各按同一排序子句取 offset+size 条，内存归并排序后截取当前页，保证跨表分页正确。
// 点赞状态/点赞数：Redis 缓存优先 → 未命中用 DB/主表字段兜底 → 回填缓存（与详情页读写一致）。
func (s *homeService) GetFeed(memberID uint, req *requestHome.FeedRequest) (*responseHome.FeedResponse, error) {
	// ========== 参数默认值 ==========
	// page 上限与 response.PageRequest 保持一致（req 不再内嵌它，改由这里兜底）
	page, size := req.Page, req.Size
	if page < 1 {
		page = 1
	}
	if page > 10000 {
		page = 10000
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	// 空值或非法值统一兜底为「推荐」
	sortMode := req.Sort
	order, ok := feedOrderMap[sortMode]
	if !ok {
		sortMode, order = "推荐", feedOrderMap["推荐"]
	}
	category := req.Category
	if category < 0 {
		category = 0
	}

	// ========== 游标解析（须早于热门标签查询：推荐模式要用游标里的模式取热门集合） ==========
	// 客户端翻页只需把上次响应的 next_cursor 原样带回，sort/category 可以省略；
	// 游标里的模式与分类具有权威性，仅在请求显式传了且与游标冲突时判非法，
	// 用于拦截「换了排序/分类后复用旧游标」。
	var cur *repoHome.FeedCursor
	if req.Cursor != "" {
		c, valid := repoHome.DecodeFeedCursor(req.Cursor)
		if !valid {
			return nil, errors.New(errors.CodeInvalidParam, "无效的分页游标")
		}
		if req.Sort != "" && c.Mode != sortMode {
			return nil, errors.New(errors.CodeInvalidParam, "无效的分页游标")
		}
		if category > 0 && c.Category != category {
			return nil, errors.New(errors.CodeInvalidParam, "无效的分页游标")
		}
		cur = c
		sortMode, category = c.Mode, c.Category
		// 翻页不传 size 时沿用游标里的值，避免只带 cursor 时每页条数突然变成默认 10
		if req.Size <= 0 && c.Size > 0 {
			size = c.Size
		}
	}

	firstPage := cur == nil && page == 1
	if firstPage {
		if cached, ok := s.homeRepo.GetCachedFirstFeed(context.Background(), sortMode, category, size); ok {
			if err := s.attachFeedLikes(memberID, cached.List); err != nil {
				return nil, err
			}
			return cached, nil
		}
	}

	// ========== 推荐模式：热门标签内容置顶 ==========
	// 热门标签 is_hot=1 的标签 → 带这些标签的笔记/攻略 ID 集合；
	// 无热门标签时集合为空，排序退化为纯发布时间倒序（等于「最新」）。
	var hotNoteIDs, hotStrategyIDs []uint
	hotNoteSet := make(map[uint]bool)
	hotStrategySet := make(map[uint]bool)
	if sortMode == "推荐" {
		hotTagIDs, err := s.homeRepo.FindHotTagIDs()
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询热门标签失败", err)
		}
		hotNoteIDs, err = s.homeRepo.FindHotNoteIDs(hotTagIDs)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询热门笔记失败", err)
		}
		hotStrategyIDs, err = s.homeRepo.FindHotStrategyIDs(hotTagIDs)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询热门攻略失败", err)
		}
		for _, id := range hotNoteIDs {
			hotNoteSet[id] = true
		}
		for _, id := range hotStrategyIDs {
			hotStrategySet[id] = true
		}
	}

	// ========== 游标分页（keyset）：带 cursor 时走新路径 ==========
	// 游标里存了模式与分类，复用/串用已在上面判为非法，而不是给出静默错误的结果。
	if cur != nil {
		return s.getFeedKeyset(feedPageContext{
			memberID:       memberID,
			sortMode:       sortMode,
			category:       category,
			size:           size,
			hotNoteIDs:     hotNoteIDs,
			hotStrategyIDs: hotStrategyIDs,
			hotNoteSet:     hotNoteSet,
			hotStrategySet: hotStrategySet,
		}, cur)
	}

	// ========== 跨表统计总数 ==========
	totalNotes, err := s.homeRepo.CountFeedNotes(category)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "统计笔记失败", err)
	}
	totalStrategies, err := s.homeRepo.CountFeedStrategies(category)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "统计攻略失败", err)
	}
	total := totalNotes + totalStrategies
	if total == 0 {
		result := &responseHome.FeedResponse{List: []responseHome.FeedItem{}, Total: 0, TotalPage: 0, Page: page, Size: size}
		if firstPage {
			s.homeRepo.SetCachedFirstFeed(context.Background(), sortMode, category, size, result)
		}
		return result, nil
	}

	// ========== 窗口拉取（每表取 offset+size 条，足够覆盖全局第 offset 条） ==========
	offset := (page - 1) * size
	limit := offset + size

	notes, err := s.homeRepo.FindFeedNotes(category, order, limit, hotNoteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记失败", err)
	}
	strategies, err := s.homeRepo.FindFeedStrategies(category, order, limit, hotStrategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略失败", err)
	}

	// ========== 跨表合并排序 ==========
	merged := mergeFeedEntries(notes, strategies, hotNoteSet, hotStrategySet, sortMode)

	// ========== 截取当前页 ==========
	if offset >= len(merged) {
		return &responseHome.FeedResponse{List: []responseHome.FeedItem{}, Total: total, TotalPage: int(math.Ceil(float64(total) / float64(size))), Page: page, Size: size}, nil
	}
	end := offset + size
	if end > len(merged) {
		end = len(merged)
	}
	pageEntries := merged[offset:end]

	// ========== 组装 FeedItem：精简 FeedEntry → 复用 BuildFeedItems（唯一组装实现，search 模块共用） ==========
	feedEntries := make([]FeedEntry, len(pageEntries))
	for i, e := range pageEntries {
		feedEntries[i] = FeedEntry{
			IsNote: e.isNote, ID: e.id, MemberID: e.memberID, Title: e.title,
			VideoURL: e.videoURL, Category: e.category, LikeCount: e.likeCount,
		}
	}
	items, err := s.buildFeedContent(feedEntries)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "组装Feed失败", err)
	}

	// has_more 由总数判定：offset+size 已覆盖到 total 即无下一页
	hasMore := offset+size < int(total)

	// 本路径也产出游标：本页最后一条就是全局第 offset+size-1 条，正是 keyset 需要的边界元素。
	// 这样客户端从普通请求（甚至旧的 ?page=N）拿到 next_cursor 后即可无缝改用游标续翻，
	// 无需额外的开关参数，也不必先做一次"取第一页"的特殊约定。
	nextCursor := ""
	if hasMore {
		nextCursor = repoHome.EncodeFeedCursor(feedCursorFromEntry(pageEntries[len(pageEntries)-1], sortMode, category, size, page))
	}

	result := &responseHome.FeedResponse{
		List: items, Total: total, TotalPage: int(math.Ceil(float64(total) / float64(size))),
		Page: page, Size: size, NextCursor: nextCursor, HasMore: hasMore,
	}
	if firstPage {
		s.homeRepo.SetCachedFirstFeed(context.Background(), sortMode, category, size, result)
	}
	if err := s.attachFeedLikes(memberID, result.List); err != nil {
		return nil, err
	}
	return result, nil
}

// feedPageContext Feed 分页的公共上下文（offset 路径与 keyset 路径共用的查询条件与热门集合）
type feedPageContext struct {
	memberID       uint
	sortMode       string
	category       int
	size           int
	hotNoteIDs     []uint
	hotStrategyIDs []uint
	hotNoteSet     map[uint]bool
	hotStrategySet map[uint]bool
}

// getFeedKeyset 游标分页取一页
//
// 取数：每表只取 size+1 条（与页码无关）。size+1 既必要又充分——设游标后剩余笔记 R_n、攻略 R_s，
// 两表各取 min(size+1,R_*) 合并后长度为 M，则 M>size 当且仅当 R_n+R_s>size：
// 任一表剩余超过 size 时它单独就能贡献 size+1 条；两表都不足时 M 就是剩余总数。故无需 offset，
// 且合并后多出的那一条恰好能判定 has_more。
//
// total/total_page/page 仍按旧口径返回：total 必须真实，无法由游标推导，所以这里仍跑 2 次 COUNT
// （相对纯 keyset 的唯一性能让步）。
func (s *homeService) getFeedKeyset(p feedPageContext, cur *repoHome.FeedCursor) (*responseHome.FeedResponse, error) {
	limit := p.size + 1

	notes, err := s.homeRepo.FindFeedNotesKeyset(repoHome.FeedKeysetQuery{
		Mode: p.sortMode, Category: p.category, Limit: limit, HotIDs: p.hotNoteIDs, Cursor: cur,
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记失败", err)
	}
	strategies, err := s.homeRepo.FindFeedStrategiesKeyset(repoHome.FeedKeysetQuery{
		Mode: p.sortMode, Category: p.category, Limit: limit, HotIDs: p.hotStrategyIDs, Cursor: cur,
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略失败", err)
	}

	merged := mergeFeedEntries(notes, strategies, p.hotNoteSet, p.hotStrategySet, p.sortMode)

	hasMore := len(merged) > p.size
	if hasMore {
		merged = merged[:p.size]
	}

	page := cur.Page + 1

	totalNotes, err := s.homeRepo.CountFeedNotes(p.category)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "统计笔记失败", err)
	}
	totalStrategies, err := s.homeRepo.CountFeedStrategies(p.category)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "统计攻略失败", err)
	}
	total := totalNotes + totalStrategies

	// 组装 FeedItem（复用唯一组装实现）
	feedEntries := make([]FeedEntry, len(merged))
	for i, e := range merged {
		feedEntries[i] = FeedEntry{
			IsNote: e.isNote, ID: e.id, MemberID: e.memberID, Title: e.title,
			VideoURL: e.videoURL, Category: e.category, LikeCount: e.likeCount,
		}
	}
	items, err := s.BuildFeedItems(p.memberID, feedEntries)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "组装Feed失败", err)
	}

	// 游标由本页实际返回的最后一条构造（不是未消费的那一条），这是 keyset 的标准约定
	nextCursor := ""
	if hasMore && len(merged) > 0 {
		nextCursor = repoHome.EncodeFeedCursor(feedCursorFromEntry(merged[len(merged)-1], p.sortMode, p.category, p.size, page))
	}

	return &responseHome.FeedResponse{
		List: items, Total: total, TotalPage: int(math.Ceil(float64(total) / float64(p.size))),
		Page: page, Size: p.size, NextCursor: nextCursor, HasMore: hasMore,
	}, nil
}

// mergeFeedEntries 跨表合并两个来源并排序（offset 路径与 keyset 路径共用）
func mergeFeedEntries(notes []entity.Note, strategies []entity.Strategy, hotNoteSet, hotStrategySet map[uint]bool, sortMode string) []feedEntry {
	merged := make([]feedEntry, 0, len(notes)+len(strategies))
	for _, n := range notes {
		merged = append(merged, feedEntry{
			isNote: true, id: n.ID, memberID: n.MemberID, title: n.Title,
			videoURL: n.VideoURL, category: n.Category,
			likeCount: n.LikeCount, heat: int64(n.Heat), publish: n.PublishTime,
			isHot: hotNoteSet[n.ID],
		})
	}
	for _, st := range strategies {
		merged = append(merged, feedEntry{
			isNote: false, id: st.ID, memberID: st.MemberID, title: st.Title,
			videoURL: st.VideoURL, category: st.Category,
			likeCount: st.LikeCount, heat: int64(st.Heat), publish: st.PublishTime,
			isHot: hotStrategySet[st.ID],
		})
	}
	sortFeedEntries(merged, sortMode)
	return merged
}

// sortFeedEntries 跨表归并排序（offset 路径与 keyset 路径共用）
//
//	最新：publish_time DESC → id DESC → isNote DESC
//	最热：heat DESC → like_count DESC → id DESC → isNote DESC
//	推荐：isHot DESC → publish_time DESC → id DESC → isNote DESC
//
// 最后一级 isNote（笔记在攻略前）不可省略：note 与 strategy 的 (publish_time,id) 可能完全相同
// （两表各自自增，且同一批导入的内容发布时间可精确到秒相同），少了它排序不是全序，
// sort.Slice 不稳定、游标谓词也会判错，导致边界处重复或漏行。keyset 分页要求严格的唯一排序键。
func sortFeedEntries(merged []feedEntry, sortMode string) {
	sort.Slice(merged, func(i, j int) bool {
		a, b := merged[i], merged[j]
		switch sortMode {
		case "最新":
			if !a.publish.Equal(b.publish) {
				return a.publish.After(b.publish)
			}
		case "最热":
			if a.heat != b.heat {
				return a.heat > b.heat
			}
			if a.likeCount != b.likeCount {
				return a.likeCount > b.likeCount
			}
		default: // 推荐：热门置顶 → 发布时间倒序 → id 倒序
			if a.isHot != b.isHot {
				return a.isHot
			}
			if !a.publish.Equal(b.publish) {
				return a.publish.After(b.publish)
			}
		}
		if a.id != b.id {
			return a.id > b.id
		}
		return a.isNote && !b.isNote
	})
}

// feedCursorFromEntry 由边界元素构造下一页游标
func feedCursorFromEntry(e feedEntry, mode string, category, size, page int) *repoHome.FeedCursor {
	return &repoHome.FeedCursor{
		Mode: mode, Category: category, Size: size, Page: page, IsNote: e.isNote, ID: e.id,
		Publish: e.publish, Hot: int(e.heat), Like: e.likeCount, IsHot: feedHotFlag(e.isHot),
	}
}

// feedHotFlag 热门标记 → hot_flag（0=热门置顶段，1=普通段），与 SQL 的 CASE 表达式一致
func feedHotFlag(isHot bool) int {
	if isHot {
		return 0
	}
	return 1
}

// BuildFeedItems 组装 FeedItem 列表（GetFeed 与 search 共用）。
func (s *homeService) BuildFeedItems(memberID uint, entries []FeedEntry) ([]responseHome.FeedItem, error) {
	items, err := s.buildFeedContent(entries)
	if err != nil {
		return nil, err
	}
	if err := s.attachFeedLikes(memberID, items); err != nil {
		return nil, err
	}
	return items, nil
}

// buildFeedContent 只组装所有用户共享的展示字段，结果可安全放入首屏缓存。
func (s *homeService) buildFeedContent(entries []FeedEntry) ([]responseHome.FeedItem, error) {
	noteIDs, strategyIDs, memberIDs := splitFeedIDs(entries)

	noteCovers, err := s.homeRepo.FindFirstImagesByNoteIDs(noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记封面失败", err)
	}
	strategyCovers, err := s.homeRepo.FindFirstImagesByStrategyIDs(strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略封面失败", err)
	}
	noteTags, err := s.homeRepo.FindFirstTagsByNoteIDs(noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记标签失败", err)
	}
	strategyTags, err := s.homeRepo.FindFirstTagsByStrategyIDs(strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略标签失败", err)
	}
	members, err := s.homeRepo.FindMembersByIDs(memberIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询作者失败", err)
	}

	// ========== 组装（对齐契约：视频内容 images=[]，图文内容 videoUrl=null） ==========
	items := make([]responseHome.FeedItem, 0, len(entries))
	for _, e := range entries {
		item := responseHome.FeedItem{
			ID:       strconv.FormatUint(uint64(e.ID), 10),
			Title:    s.mask(e.Title),
			Images:   []string{},
			Tags:     []string{},
			Category: e.Category,
		}
		if m, ok := members[e.MemberID]; ok && m != nil {
			item.Member = &responseHome.FeedAuthor{Nickname: s.mask(m.Nickname), Avatar: m.Avatar}
		}
		if e.IsNote {
			item.Type = "note"
			if tag, ok := noteTags[e.ID]; ok {
				item.Tags = []string{tag}
			} else {
				item.Tags = []string{activityCategory[e.Category]}
			}
			if cover, ok := noteCovers[e.ID]; ok {
				item.Images = []string{cover}
			}
			if e.VideoURL != "" {
				item.VideoURL = &e.VideoURL
			}
		} else {
			item.Type = "guide"
			if tag, ok := strategyTags[e.ID]; ok {
				item.Tags = []string{tag}
			} else {
				item.Tags = []string{activityCategory[e.Category]}
			}
			if cover, ok := strategyCovers[e.ID]; ok {
				item.Images = []string{cover}
			}
			if e.VideoURL != "" {
				item.VideoURL = &e.VideoURL
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *homeService) attachFeedLikes(memberID uint, items []responseHome.FeedItem) error {
	var noteIDs, strategyIDs []uint
	for _, item := range items {
		id, err := strconv.ParseUint(item.ID, 10, 64)
		if err != nil {
			return errors.NewWithErr(errors.CodeInternalError, "Feed内容ID无效", err)
		}
		if item.Type == "note" {
			noteIDs = append(noteIDs, uint(id))
		} else {
			strategyIDs = append(strategyIDs, uint(id))
		}
	}
	ctx := context.Background()
	noteLikes, err := s.resolveNoteLikes(ctx, noteIDs, memberID)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "查询笔记点赞状态失败", err)
	}
	strategyLikes, err := s.resolveStrategyLikes(ctx, strategyIDs, memberID)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "查询攻略点赞状态失败", err)
	}
	noteCounts, err := s.resolveNoteLikeCounts(ctx, noteIDs)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "查询笔记点赞数失败", err)
	}
	strategyCounts, err := s.resolveStrategyLikeCounts(ctx, strategyIDs)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "查询攻略点赞数失败", err)
	}
	for i := range items {
		id, _ := strconv.ParseUint(items[i].ID, 10, 64)
		if items[i].Type == "note" {
			items[i].IsLiked = noteLikes[uint(id)]
			items[i].LikeCount = noteCounts[uint(id)]
		} else {
			items[i].IsLiked = strategyLikes[uint(id)]
			items[i].LikeCount = strategyCounts[uint(id)]
		}
	}
	return nil
}

// resolveNoteLikes 笔记点赞状态：Redis 缓存优先，未命中批量查 DB 并回填缓存
func (s *homeService) resolveNoteLikes(ctx context.Context, noteIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(noteIDs))
	var missing []uint
	for _, id := range noteIDs {
		isLiked, found, err := s.noteRepo.GetCachedLikeStatus(ctx, id, memberID)
		if err == nil && found {
			result[id] = isLiked
		} else {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return result, nil
	}
	dbStatuses, err := s.homeRepo.FindNoteLikeStatuses(missing, memberID)
	if err != nil {
		return nil, err
	}
	for _, id := range missing {
		isLiked := dbStatuses[id]
		result[id] = isLiked
		s.noteRepo.SetCachedLikeStatus(ctx, id, memberID, isLiked)
	}
	return result, nil
}

// resolveStrategyLikes 攻略点赞状态：Redis 缓存优先，未命中批量查 DB 并回填缓存
func (s *homeService) resolveStrategyLikes(ctx context.Context, strategyIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(strategyIDs))
	var missing []uint
	for _, id := range strategyIDs {
		isLiked, found, err := s.guideRepo.GetCachedLikeStatus(ctx, id, memberID)
		if err == nil && found {
			result[id] = isLiked
		} else {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return result, nil
	}
	dbStatuses, err := s.homeRepo.FindStrategyLikeStatuses(missing, memberID)
	if err != nil {
		return nil, err
	}
	for _, id := range missing {
		isLiked := dbStatuses[id]
		result[id] = isLiked
		s.guideRepo.SetCachedLikeStatus(ctx, id, memberID, isLiked)
	}
	return result, nil
}

// resolveNoteLikeCounts 笔记点赞数：Redis 优先，未命中批量查询主表。
func (s *homeService) resolveNoteLikeCounts(ctx context.Context, noteIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(noteIDs))
	var missing []uint
	for _, id := range noteIDs {
		count, err := s.noteRepo.GetCachedLikeCount(ctx, id)
		if err == nil {
			result[id] = count
			continue
		}
		missing = append(missing, id)
	}
	counts, err := s.homeRepo.FindNoteLikeCounts(missing)
	if err != nil {
		return nil, err
	}
	for _, id := range missing {
		result[id] = counts[id]
		s.noteRepo.SetCachedLikeCount(ctx, id, counts[id])
	}
	return result, nil
}

// resolveStrategyLikeCounts 攻略点赞数：Redis 优先，未命中批量查询主表。
func (s *homeService) resolveStrategyLikeCounts(ctx context.Context, strategyIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(strategyIDs))
	var missing []uint
	for _, id := range strategyIDs {
		count, err := s.guideRepo.GetCachedLikeCount(ctx, id)
		if err == nil {
			result[id] = count
			continue
		}
		missing = append(missing, id)
	}
	counts, err := s.homeRepo.FindStrategyLikeCounts(missing)
	if err != nil {
		return nil, err
	}
	for _, id := range missing {
		result[id] = counts[id]
		s.guideRepo.SetCachedLikeCount(ctx, id, counts[id])
	}
	return result, nil
}

// splitFeedIDs 拆分条目的笔记ID、攻略ID、作者ID（去重）
func splitFeedIDs(entries []FeedEntry) (noteIDs, strategyIDs, memberIDs []uint) {
	noteIDs = make([]uint, 0, len(entries))
	strategyIDs = make([]uint, 0, len(entries))
	memberIDs = make([]uint, 0, len(entries))
	for _, e := range entries {
		memberIDs = append(memberIDs, e.MemberID)
		if e.IsNote {
			noteIDs = append(noteIDs, e.ID)
		} else {
			strategyIDs = append(strategyIDs, e.ID)
		}
	}
	return dedupeUint(noteIDs), dedupeUint(strategyIDs), dedupeUint(memberIDs)
}

func dedupeUint(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// activityCategory 活动类型 → 中文分类
var activityCategory = map[int]string{
	1: "徒步", 2: "露营", 3: "骑行", 4: "登山", 5: "跑步", 6: "攀岩", 7: "溯溪",
}

// feedOrderMap 排序模式 → SQL 排序子句（含 id 兜底，保证跨表分页窗口稳定）
// 推荐：热门置顶 CASE 由仓储在 hotIDs 非空时动态拼接，此处为基础时间序；无热门时即等于「最新」。
var feedOrderMap = map[string]string{
	"推荐": "publish_time DESC, id DESC",
	"最新": "publish_time DESC, id DESC",
	"最热": "heat DESC, like_count DESC, id DESC",
}

// feedEntry 跨表归并排序用的中间结构（home 私有：带推荐/最热排序字段）
type feedEntry struct {
	isNote    bool
	isHot     bool // 推荐模式：是否带热门标签（is_hot=1），热门整体置顶
	id        uint
	memberID  uint
	title     string
	videoURL  string
	category  int
	likeCount int64
	heat      int64
	publish   time.Time
}

// FeedEntry 精简跨表条目（导出，GetFeed 与 search 模块组装 FeedItem 共用）。
// 只携带展示所需字段，不带 home 排序用的 isHot/heat/publish。
type FeedEntry struct {
	IsNote    bool   // true=笔记，false=攻略
	ID        uint   // 笔记ID或攻略ID
	MemberID  uint   // 作者ID
	Title     string // 标题
	VideoURL  string // 视频URL（图文为空）
	Category  int    // 分类枚举
	LikeCount int64  // 主表点赞数字段（点赞数缓存未命中时兜底）
}

// mask 前台文本出口统一打星（nil 安全）
func (s *homeService) mask(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}
