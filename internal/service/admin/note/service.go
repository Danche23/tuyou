package note

import (
	"context"
	"fmt"
	"math"

	request "tuyou/internal/model/dto/request/admin/note"
	response "tuyou/internal/model/dto/response/admin/note"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/note"
	"tuyou/pkg/errors"
)

type noteService struct {
	noteRepo repo.NoteRepository
}

// NewNoteService 创建后台笔记服务实例
func NewNoteService(noteRepo repo.NoteRepository) NoteService {
	return &noteService{noteRepo: noteRepo}
}

// ListNotes 分页查询笔记列表
func (s *noteService) ListNotes(ctx context.Context, req *request.ListNoteReq) (*response.ListNoteResp, error) {
	// ========== 1. 拼装筛选条件 ==========
	filters := make(map[string]interface{})

	if req.Title != "" {
		filters["title"] = req.Title
	}
	if req.Category != "" {
		c, ok := categoryStrToInt[req.Category]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "分类不合法")
		}
		filters["category"] = c
	}
	if req.IsSlide != "" {
		switch req.IsSlide {
		case "是":
			filters["is_slide"] = 1
		case "否":
			filters["is_slide"] = 0
		default:
			return nil, errors.New(errors.CodeInvalidParam, "是否轮播参数不合法")
		}
	}
	if req.BeginTime != "" {
		filters["begin_time"] = req.BeginTime
	}
	if req.FinishTime != "" {
		filters["finish_time"] = req.FinishTime + " 23:59:59"
	}

	// ========== 2. 排序解析 ==========
	sortField := noteSortFieldMap[req.SortField] // 不在白名单 → 空串（仓储默认 id DESC）
	sortOrder := "DESC"
	if req.SortOrder != "" {
		switch req.SortOrder {
		case "升序":
			sortOrder = "ASC"
		case "降序":
			sortOrder = "DESC"
		default:
			return nil, errors.New(errors.CodeInvalidParam, "排序方式不合法")
		}
	}

	// ========== 3. 查询总数 + 本页笔记 ==========
	total, err := s.noteRepo.Count(ctx, filters)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取笔记列表失败", err)
	}

	notes, err := s.noteRepo.FindList(ctx, filters, sortField, sortOrder, req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取笔记列表失败", err)
	}

	list := make([]response.NoteListItem, 0, len(notes))
	if len(notes) == 0 {
		return s.buildResp(list, total, req), nil
	}

	// ========== 4. 批量查询作者 / 标签 / 封面 ==========
	noteIDs := make([]uint, len(notes))
	uniqueMemberIDs := make([]uint, 0)
	memberIDSet := make(map[uint]bool)
	for i, n := range notes {
		noteIDs[i] = n.ID
		if !memberIDSet[n.MemberID] {
			memberIDSet[n.MemberID] = true
			uniqueMemberIDs = append(uniqueMemberIDs, n.MemberID)
		}
	}

	members, err := s.noteRepo.FindMembersByIDs(ctx, uniqueMemberIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询作者信息失败", err)
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}

	tagsMap, err := s.noteRepo.FindTagsByNoteIDs(ctx, noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}

	covers, err := s.noteRepo.FindCoverImagesByNoteIDs(ctx, noteIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询封面失败", err)
	}

	// ========== 5. 组装响应 ==========
	for _, n := range notes {
		var nickname string
		if m, ok := memberMap[n.MemberID]; ok {
			nickname = m.Nickname
		}

		category := categoryIntToStr[n.Category]
		if category == "" {
			category = "未知"
		}

		isSlide := "否"
		if n.IsSlide == 1 {
			isSlide = "是"
		}

		status := "下架"
		if n.Status == 1 {
			status = "正常"
		}

		list = append(list, response.NoteListItem{
			ID:           fmt.Sprintf("N%04d", n.ID),
			CoverImg:     covers[n.ID],
			Title:        n.Title,
			Nickname:     nickname,
			Category:     category,
			Tags:         tagsMap[n.ID],
			PublishTime:  n.PublishTime.Format("2006-01-02 15:04:05"),
			ViewCount:    n.ViewCount,
			LikeCount:    n.LikeCount,
			CollectCount: n.CollectCount,
			CommentCount: n.CommentCount,
			Heat:         int64(n.Heat),
			IsSlide:      isSlide,
			Status:       status,
		})
	}

	return s.buildResp(list, total, req), nil
}

// buildResp 组装分页响应
func (s *noteService) buildResp(list []response.NoteListItem, total int64, req *request.ListNoteReq) *response.ListNoteResp {
	return &response.ListNoteResp{
		List:      list,
		Total:     total,
		Page:      req.Page,
		Size:      req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
	}
}

// 分类数字 → 中文
var categoryIntToStr = map[int]string{
	1: "徒步", 2: "露营", 3: "骑行", 4: "登山", 5: "跑步", 6: "攀岩", 7: "溯溪",
}

// 分类中文 → 数字
var categoryStrToInt = map[string]int{
	"徒步": 1, "露营": 2, "骑行": 3, "登山": 4, "跑步": 5, "攀岩": 6, "溯溪": 7,
}

// 排序字段中文 → 数据库列名（白名单，防注入）
var noteSortFieldMap = map[string]string{
	"热度":   "heat",
	"浏览量": "view_count",
	"点赞量": "like_count",
	"收藏量": "collect_count",
	"评论量": "comment_count",
}
