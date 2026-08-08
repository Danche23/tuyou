package home

import (
	"context"
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
)

type homeService struct {
	homeRepo  repoHome.HomeRepository
	noteRepo  repoNote.NoteRepository
	guideRepo repoGuide.GuideRepository
}

// NewHomeService 创建首页服务实例
// 注入 note/guide 仓储以复用点赞状态缓存（Redis 优先）
func NewHomeService(homeRepo repoHome.HomeRepository, noteRepo repoNote.NoteRepository, guideRepo repoGuide.GuideRepository) HomeService {
	return &homeService{homeRepo: homeRepo, noteRepo: noteRepo, guideRepo: guideRepo}
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
		items = append(items, toActivitySlide(act))
	}
	for _, n := range slideNotes {
		items = append(items, toNoteSlide(n, noteCovers[n.ID], noteTags[n.ID]))
	}
	for _, st := range slideStrategies {
		items = append(items, toStrategySlide(st, strategyCovers[st.ID], strategyTags[st.ID]))
	}
	for _, act := range fillActs {
		items = append(items, toActivitySlide(act))
	}
	for _, n := range fillNotes {
		items = append(items, toNoteSlide(n, noteCovers[n.ID], noteTags[n.ID]))
	}
	for _, st := range fillStrategies {
		items = append(items, toStrategySlide(st, strategyCovers[st.ID], strategyTags[st.ID]))
	}

	// 防御性截断，最多 5 张
	if len(items) > 5 {
		items = items[:5]
	}
	return items, nil
}

// toActivitySlide 活动 → 轮播项（字段全部来自活动主表）
func toActivitySlide(act entity.Activity) responseHome.SlideItem {
	return responseHome.SlideItem{
		ID:            strconv.FormatUint(uint64(act.ID), 10),
		Type:          "activity",
		Title:         act.Title,
		Cover:         act.CoverImg,
		Tag:           activityCategory[act.ActivityType],
		SignUpStart:   act.SignUpStart.Format("2006-01-02 15:04:05"),
		CurrentPeople: act.CurrentPeople,
		MaxPeople:     act.MaxPeople,
	}
}

// toNoteSlide 笔记 → 轮播项（封面/标签来自批量查询结果）
func toNoteSlide(n entity.Note, cover, tag string) responseHome.SlideItem {
	return responseHome.SlideItem{
		ID:    strconv.FormatUint(uint64(n.ID), 10),
		Type:  "note",
		Title: n.Title,
		Cover: cover,
		Tag:   tag,
	}
}

// toStrategySlide 攻略 → 轮播项（封面/标签来自批量查询结果）
func toStrategySlide(st entity.Strategy, cover, tag string) responseHome.SlideItem {
	return responseHome.SlideItem{
		ID:    strconv.FormatUint(uint64(st.ID), 10),
		Type:  "guide",
		Title: st.Title,
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
// 排序：推荐=点赞数倒序→热度倒序→发布时间倒序；最新=发布时间倒序；最热=热度倒序→点赞数倒序。
// 分页：笔记/攻略各按同一排序子句取 offset+size 条，内存归并排序后截取当前页，保证跨表分页正确。
// 点赞状态/点赞数：Redis 缓存优先 → 未命中用 DB/主表字段兜底 → 回填缓存（与详情页读写一致）。
func (s *homeService) GetFeed(memberID uint, req *requestHome.FeedRequest) (*responseHome.FeedResponse, error) {
	ctx := context.Background()

	// ========== 参数默认值 ==========
	page, size := req.Page, req.Size
	if page < 1 {
		page = 1
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
		return &responseHome.FeedResponse{List: []responseHome.FeedItem{}, Total: 0, Page: page, Size: size}, nil
	}

	// ========== 窗口拉取（每表取 offset+size 条，足够覆盖全局第 offset 条） ==========
	offset := (page - 1) * size
	limit := offset + size

	notes, err := s.homeRepo.FindFeedNotes(category, order, limit)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记失败", err)
	}
	strategies, err := s.homeRepo.FindFeedStrategies(category, order, limit)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略失败", err)
	}

	// ========== 跨表合并排序 ==========
	merged := make([]feedEntry, 0, len(notes)+len(strategies))
	for _, n := range notes {
		merged = append(merged, feedEntry{
			isNote: true, id: n.ID, memberID: n.MemberID, title: n.Title,
			videoURL: n.VideoURL, category: n.Category,
			likeCount: n.LikeCount, heat: int64(n.Heat), publish: n.PublishTime,
		})
	}
	for _, st := range strategies {
		merged = append(merged, feedEntry{
			isNote: false, id: st.ID, memberID: st.MemberID, title: st.Title,
			videoURL: st.VideoURL, category: st.Category,
			likeCount: st.LikeCount, heat: int64(st.Heat), publish: st.PublishTime,
		})
	}
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
		default: // 推荐
			if a.likeCount != b.likeCount {
				return a.likeCount > b.likeCount
			}
			if a.heat != b.heat {
				return a.heat > b.heat
			}
			if !a.publish.Equal(b.publish) {
				return a.publish.After(b.publish)
			}
		}
		return a.id > b.id
	})

	// ========== 截取当前页 ==========
	if offset >= len(merged) {
		return &responseHome.FeedResponse{List: []responseHome.FeedItem{}, Total: total, Page: page, Size: size}, nil
	}
	end := offset + size
	if end > len(merged) {
		end = len(merged)
	}
	pageEntries := merged[offset:end]

	// ========== 批量取关联数据（避免 N+1） ==========
	noteIDs, strategyIDs, memberIDs := splitFeedIDs(pageEntries)

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

	// 点赞状态：Redis 优先 → 批量 DB 兜底 → 回填缓存
	noteLikes, err := s.resolveNoteLikes(ctx, noteIDs, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询笔记点赞状态失败", err)
	}
	strategyLikes, err := s.resolveStrategyLikes(ctx, strategyIDs, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询攻略点赞状态失败", err)
	}

	// ========== 点赞数：Redis 缓存优先 → 主表字段兜底 → 回填缓存 ==========
	noteCountFallback := make(map[uint]int64, len(noteIDs))
	strategyCountFallback := make(map[uint]int64, len(strategyIDs))
	for _, e := range pageEntries {
		if e.isNote {
			noteCountFallback[e.id] = e.likeCount
		} else {
			strategyCountFallback[e.id] = e.likeCount
		}
	}
	noteCounts := s.resolveNoteLikeCounts(ctx, noteIDs, noteCountFallback)
	strategyCounts := s.resolveStrategyLikeCounts(ctx, strategyIDs, strategyCountFallback)

	// ========== 组装（对齐契约：视频内容 images=[]，图文内容 videoUrl=null） ==========
	items := make([]responseHome.FeedItem, 0, len(pageEntries))
	for _, e := range pageEntries {
		item := responseHome.FeedItem{
			ID:        strconv.FormatUint(uint64(e.id), 10),
			Title:     e.title,
			Images:    []string{},
			Tags:      []string{},
			LikeCount: e.likeCount,
			Category:  e.category,
		}
		if m, ok := members[e.memberID]; ok && m != nil {
			item.Member = &responseHome.FeedAuthor{Nickname: m.Nickname, Avatar: m.Avatar}
		}
		if e.isNote {
			item.Type = "note"
			item.IsLiked = noteLikes[e.id]
			item.LikeCount = noteCounts[e.id]
			if tag, ok := noteTags[e.id]; ok {
				item.Tags = []string{tag}
			}
			if cover, ok := noteCovers[e.id]; ok {
				item.Images = []string{cover}
			}
			if e.videoURL != "" {
				item.VideoURL = &e.videoURL
			}
		} else {
			item.Type = "guide"
			item.IsLiked = strategyLikes[e.id]
			item.LikeCount = strategyCounts[e.id]
			if tag, ok := strategyTags[e.id]; ok {
				item.Tags = []string{tag}
			}
			if cover, ok := strategyCovers[e.id]; ok {
				item.Images = []string{cover}
			}
			if e.videoURL != "" {
				item.VideoURL = &e.videoURL
			}
		}
		items = append(items, item)
	}

	return &responseHome.FeedResponse{List: items, Total: total, Page: page, Size: size}, nil
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

// resolveNoteLikeCounts 笔记点赞数：Redis 缓存优先，未命中用主表字段兜底并回填缓存
func (s *homeService) resolveNoteLikeCounts(ctx context.Context, noteIDs []uint, fallbacks map[uint]int64) map[uint]int64 {
	result := make(map[uint]int64, len(noteIDs))
	for _, id := range noteIDs {
		count, err := s.noteRepo.GetCachedLikeCount(ctx, id)
		if err == nil {
			result[id] = count
			continue
		}
		fb := fallbacks[id]
		s.noteRepo.SetCachedLikeCount(ctx, id, fb)
		result[id] = fb
	}
	return result
}

// resolveStrategyLikeCounts 攻略点赞数：Redis 缓存优先，未命中用主表字段兜底并回填缓存
func (s *homeService) resolveStrategyLikeCounts(ctx context.Context, strategyIDs []uint, fallbacks map[uint]int64) map[uint]int64 {
	result := make(map[uint]int64, len(strategyIDs))
	for _, id := range strategyIDs {
		count, err := s.guideRepo.GetCachedLikeCount(ctx, id)
		if err == nil {
			result[id] = count
			continue
		}
		fb := fallbacks[id]
		s.guideRepo.SetCachedLikeCount(ctx, id, fb)
		result[id] = fb
	}
	return result
}

// splitFeedIDs 拆分当前页条目的笔记ID、攻略ID、作者ID（去重）
func splitFeedIDs(entries []feedEntry) (noteIDs, strategyIDs, memberIDs []uint) {
	noteIDs = make([]uint, 0, len(entries))
	strategyIDs = make([]uint, 0, len(entries))
	memberIDs = make([]uint, 0, len(entries))
	for _, e := range entries {
		memberIDs = append(memberIDs, e.memberID)
		if e.isNote {
			noteIDs = append(noteIDs, e.id)
		} else {
			strategyIDs = append(strategyIDs, e.id)
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
var feedOrderMap = map[string]string{
	"推荐": "like_count DESC, heat DESC, publish_time DESC, id DESC",
	"最新": "publish_time DESC, id DESC",
	"最热": "heat DESC, like_count DESC, id DESC",
}

// feedEntry 跨表归并排序用的中间结构
type feedEntry struct {
	isNote    bool
	id        uint
	memberID  uint
	title     string
	videoURL  string
	category  int
	likeCount int64
	heat      int64
	publish   time.Time
}
