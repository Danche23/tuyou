package guide

import (
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/guide"
	response "tuyou/internal/model/dto/response/admin/guide"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/guide"
	"tuyou/pkg/errors"
)

type guideService struct {
	guideRepo repo.GuideRepository
	db        *gorm.DB
}

// NewGuideService 创建后台攻略服务实例
func NewGuideService(guideRepo repo.GuideRepository, db *gorm.DB) GuideService {
	return &guideService{guideRepo: guideRepo, db: db}
}

// ListGuides 分页查询攻略列表
func (s *guideService) ListGuides(ctx context.Context, req *request.ListGuideReq) (*response.ListGuideResp, error) {
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
	if req.Difficulty != "" {
		d, ok := difficultyStrToInt[req.Difficulty]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "难度不合法")
		}
		filters["difficulty"] = d
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
	sortField := guideSortFieldMap[req.SortField] // 不在白名单 → 空串（仓储默认 id DESC）
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

	// ========== 3. 查询总数 + 本页攻略 ==========
	total, err := s.guideRepo.Count(ctx, filters)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取攻略列表失败", err)
	}

	strategies, err := s.guideRepo.FindList(ctx, filters, sortField, sortOrder, req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取攻略列表失败", err)
	}

	list, err := s.buildListItems(ctx, strategies)
	if err != nil {
		return nil, err
	}

	return s.buildResp(list, total, req), nil
}

// GetGuideDetail 获取攻略详情
func (s *guideService) GetGuideDetail(ctx context.Context, id uint) (*response.GuideDetailResp, error) {
	g, err := s.guideRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取攻略详情失败", err)
	}
	if g == nil {
		return nil, errors.New(errors.CodeNotFound, "攻略不存在")
	}

	// 作者昵称
	members, err := s.guideRepo.FindMembersByIDs(ctx, []uint{g.MemberID})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询作者信息失败", err)
	}
	nickname := ""
	if len(members) > 0 {
		nickname = members[0].Nickname
	}

	// 标签
	tagsMap, err := s.guideRepo.FindTagsByStrategyIDs(ctx, []uint{g.ID})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}

	// 封面
	covers, err := s.guideRepo.FindCoverImagesByStrategyIDs(ctx, []uint{g.ID})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询封面失败", err)
	}

	category := categoryIntToStr[g.Category]
	if category == "" {
		category = "未知"
	}

	difficulty := difficultyIntToStr[g.Difficulty]
	if difficulty == "" {
		difficulty = "未知"
	}

	isSlide := "否"
	if g.IsSlide == 1 {
		isSlide = "是"
	}

	status := "下架"
	if g.Status == 1 {
		status = "正常"
	}

	deleteTime := ""
	if g.DeletedAt.Valid {
		deleteTime = g.DeletedAt.Time.Format("2006-01-02 15:04:05")
	}

	return &response.GuideDetailResp{
		GuideListItem: response.GuideListItem{
			ID:           fmt.Sprintf("G%04d", g.ID),
			CoverImg:     covers[g.ID],
			Title:        g.Title,
			Nickname:     nickname,
			Category:     category,
			Tags:         tagsMap[g.ID],
			Difficulty:   difficulty,
			PublishTime:  g.PublishTime.Format("2006-01-02 15:04:05"),
			ViewCount:    g.ViewCount,
			LikeCount:    g.LikeCount,
			CollectCount: g.CollectCount,
			CommentCount: g.CommentCount,
			Heat:         int64(g.Heat),
			IsSlide:      isSlide,
			Status:       status,
		},
		Content:    g.Content,
		Distance:   g.Distance,
		Duration:   g.Duration,
		Transport:  g.Transport,
		VideoURL:   g.VideoURL,
		DeleteTime: deleteTime,
	}, nil
}

// UpdateGuide 编辑攻略（分类/轮播/状态三字段必传，整体覆盖），返回更新后的详情
func (s *guideService) UpdateGuide(ctx context.Context, id uint, req *request.UpdateGuideReq) (*response.GuideDetailResp, error) {
	// 1. 校验攻略存在
	g, err := s.guideRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取攻略失败", err)
	}
	if g == nil {
		return nil, errors.New(errors.CodeNotFound, "攻略不存在")
	}

	// 2. 解析三个必传字段
	c, ok := categoryStrToInt[req.Category]
	if !ok {
		return nil, errors.New(errors.CodeInvalidParam, "分类不合法")
	}

	var isSlide int
	switch req.IsSlide {
	case "是":
		isSlide = 1
	case "否":
		isSlide = 0
	default:
		return nil, errors.New(errors.CodeInvalidParam, "是否轮播参数不合法")
	}

	var status int
	switch req.Status {
	case "正常":
		status = 1
	case "下架":
		status = 0
	default:
		return nil, errors.New(errors.CodeInvalidParam, "状态参数不合法")
	}

	// 3. 事务内：更新攻略 + 状态变化时同步评论状态与发布数
	err = s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"category": c,
			"is_slide": isSlide,
			"status":   status,
		}
		// 下架统一关联 delete_time（前后台一致）：下架写删除时间，恢复正常清除
		if g.Status != status {
			if status == 0 {
				updates["delete_time"] = time.Now()
			} else {
				updates["delete_time"] = nil
			}
		}
		if err := s.guideRepo.Update(tx, id, updates); err != nil {
			return err
		}

		// 状态变化才同步评论状态与发布数（攻略评论 target_type=2）
		if g.Status != status {
			if err := s.guideRepo.UpdateCommentStatusByTarget(tx, 2, id, status); err != nil {
				return err
			}
			if status == 0 { // 正常→下架：发布数 -1
				return s.guideRepo.DecrPublishCount(tx, g.MemberID)
			}
			return s.guideRepo.IncrPublishCount(tx, g.MemberID) // 下架→正常：发布数 +1
		}
		return nil
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "编辑攻略失败", err)
	}

	// 4. 返回更新后的详情
	return s.GetGuideDetail(ctx, id)
}

// buildListItems 批量查询作者/标签/封面并组装列表项
func (s *guideService) buildListItems(ctx context.Context, strategies []entity.Strategy) ([]response.GuideListItem, error) {
	list := make([]response.GuideListItem, 0, len(strategies))
	if len(strategies) == 0 {
		return list, nil
	}

	// 批量查询作者 / 标签 / 封面
	strategyIDs := make([]uint, len(strategies))
	uniqueMemberIDs := make([]uint, 0)
	memberIDSet := make(map[uint]bool)
	for i, g := range strategies {
		strategyIDs[i] = g.ID
		if !memberIDSet[g.MemberID] {
			memberIDSet[g.MemberID] = true
			uniqueMemberIDs = append(uniqueMemberIDs, g.MemberID)
		}
	}

	members, err := s.guideRepo.FindMembersByIDs(ctx, uniqueMemberIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询作者信息失败", err)
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}

	tagsMap, err := s.guideRepo.FindTagsByStrategyIDs(ctx, strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}

	covers, err := s.guideRepo.FindCoverImagesByStrategyIDs(ctx, strategyIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询封面失败", err)
	}

	// 组装列表项
	for _, g := range strategies {
		var nickname string
		if m, ok := memberMap[g.MemberID]; ok {
			nickname = m.Nickname
		}

		category := categoryIntToStr[g.Category]
		if category == "" {
			category = "未知"
		}

		difficulty := difficultyIntToStr[g.Difficulty]
		if difficulty == "" {
			difficulty = "未知"
		}

		isSlide := "否"
		if g.IsSlide == 1 {
			isSlide = "是"
		}

		status := "下架"
		if g.Status == 1 {
			status = "正常"
		}

		list = append(list, response.GuideListItem{
			ID:           fmt.Sprintf("G%04d", g.ID),
			CoverImg:     covers[g.ID],
			Title:        g.Title,
			Nickname:     nickname,
			Category:     category,
			Tags:         tagsMap[g.ID],
			Difficulty:   difficulty,
			PublishTime:  g.PublishTime.Format("2006-01-02 15:04:05"),
			ViewCount:    g.ViewCount,
			LikeCount:    g.LikeCount,
			CollectCount: g.CollectCount,
			CommentCount: g.CommentCount,
			Heat:         int64(g.Heat),
			IsSlide:      isSlide,
			Status:       status,
		})
	}

	return list, nil
}

// buildResp 组装分页响应
func (s *guideService) buildResp(list []response.GuideListItem, total int64, req *request.ListGuideReq) *response.ListGuideResp {
	return &response.ListGuideResp{
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

// 难度中文 → 数字
var difficultyStrToInt = map[string]int{"入门": 1, "中等": 2, "中高等": 3, "高等": 4}

// 难度数字 → 中文
var difficultyIntToStr = map[int]string{1: "入门", 2: "中等", 3: "中高等", 4: "高等"}

// 排序字段中文 → 数据库列名（白名单，防注入）
var guideSortFieldMap = map[string]string{
	"热度":   "heat",
	"浏览量":  "view_count",
	"点赞量":  "like_count",
	"收藏量":  "collect_count",
	"评论量":  "comment_count",
	"发布时间": "publish_time",
}
