package home

import (
	"strconv"

	responseHome "tuyou/internal/model/dto/response/web/home"
	"tuyou/internal/model/entity"
	repoHome "tuyou/internal/repository/web/home"
	"tuyou/pkg/errors"
)

// activityCategory 活动类型 → 中文分类
var activityCategory = map[int]string{
	1: "徒步", 2: "露营", 3: "骑行", 4: "登山", 5: "跑步", 6: "攀岩", 7: "溯溪",
}

type homeService struct {
	homeRepo repoHome.HomeRepository
}

func NewHomeService(homeRepo repoHome.HomeRepository) HomeService {
	return &homeService{homeRepo: homeRepo}
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
