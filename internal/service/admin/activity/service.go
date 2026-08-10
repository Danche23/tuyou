package activity

import (
	"context"
	"fmt"
	"math"

	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/activity"
	response "tuyou/internal/model/dto/response/admin/activity"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/activity"
	"tuyou/pkg/errors"
)

type activityService struct {
	activityRepo repo.ActivityRepository
	db           *gorm.DB
}

// NewActivityService 创建后台活动服务实例
func NewActivityService(activityRepo repo.ActivityRepository, db *gorm.DB) ActivityService {
	return &activityService{activityRepo: activityRepo, db: db}
}

// ListActivities 分页查询活动列表
func (s *activityService) ListActivities(ctx context.Context, req *request.ListActivityReq) (*response.ListActivityResp, error) {
	// ========== 1. 拼装筛选条件 ==========
	filters := make(map[string]interface{})

	if req.Title != "" {
		filters["title"] = req.Title
	}
	if req.LeaderName != "" {
		filters["leader_name"] = req.LeaderName
	}
	if req.ActivityType != "" {
		t, ok := activityTypeStrToInt[req.ActivityType]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "活动类型不合法")
		}
		filters["activity_type"] = t
	}
	if req.Difficulty != "" {
		d, ok := difficultyStrToInt[req.Difficulty]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "难度不合法")
		}
		filters["difficulty"] = d
	}
	if req.Fee != "" {
		switch req.Fee {
		case "免费":
			filters["fee"] = 0
		case "付费":
			filters["fee"] = 1
		default:
			return nil, errors.New(errors.CodeInvalidParam, "费用类型不合法")
		}
	}
	if req.ActivityStatus != "" {
		st, ok := activityStatusStrToInt[req.ActivityStatus]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "活动状态不合法")
		}
		filters["activity_status"] = st
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
		filters["finish_time"] = req.FinishTime
	}

	// ========== 2. 查询总数 + 本页活动 ==========
	total, err := s.activityRepo.Count(ctx, filters)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取活动列表失败", err)
	}

	activities, err := s.activityRepo.FindList(ctx, filters, "", "", req.Page, req.Size)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取活动列表失败", err)
	}

	// ========== 3. 组装列表项 ==========
	list, err := s.buildListItems(ctx, activities)
	if err != nil {
		return nil, err
	}

	return s.buildResp(list, total, req), nil
}

// buildListItems 批量查询领队并组装列表项
func (s *activityService) buildListItems(ctx context.Context, activities []entity.Activity) ([]response.ActivityListItem, error) {
	list := make([]response.ActivityListItem, 0, len(activities))
	if len(activities) == 0 {
		return list, nil
	}

	// 批量查询领队信息
	leaderIDs := make([]uint, 0, len(activities))
	leaderSet := make(map[uint]bool)
	for _, a := range activities {
		if !leaderSet[a.LeaderMemberID] {
			leaderSet[a.LeaderMemberID] = true
			leaderIDs = append(leaderIDs, a.LeaderMemberID)
		}
	}
	members, err := s.activityRepo.FindMembersByIDs(ctx, leaderIDs)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询领队信息失败", err)
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}

	for _, a := range activities {
		var leaderName string
		if m, ok := memberMap[a.LeaderMemberID]; ok {
			leaderName = m.Nickname
		}

		isSlide := "否"
		if a.IsSlide == 1 {
			isSlide = "是"
		}

		activityStatus := activityStatusIntToStr[a.ActivityStatus]
		if activityStatus == "" {
			activityStatus = "未知"
		}

		// is_del：0→正常，1→下架
		isDel := "正常"
		if a.IsDel == 1 {
			isDel = "下架"
		}

		var deleteTime *string
		if a.DeletedAt.Valid {
			t := a.DeletedAt.Time.Format("2006-01-02 15:04:05")
			deleteTime = &t
		}

		list = append(list, response.ActivityListItem{
			ID:             fmt.Sprintf("A%03d", a.ID),
			CoverImg:       a.CoverImg,
			Title:          a.Title,
			LeaderMemberID: fmt.Sprintf("M%03d", a.LeaderMemberID),
			LeaderName:     leaderName,
			ActivityType:   activityTypeIntToStr[a.ActivityType],
			StartTime:      a.StartTime.Format("2006-01-02"),
			EndTime:        a.EndTime.Format("2006-01-02"),
			Price:          a.Price,
			CurrentPeople:  a.CurrentPeople,
			MaxPeople:      a.MaxPeople,
			Difficulty:     difficultyIntToStr[a.Difficulty],
			IsSlide:        isSlide,
			ActivityStatus: activityStatus,
			IsDel:          isDel,
			DeleteTime:     deleteTime,
		})
	}

	return list, nil
}

// buildResp 组装分页响应
func (s *activityService) buildResp(list []response.ActivityListItem, total int64, req *request.ListActivityReq) *response.ListActivityResp {
	return &response.ListActivityResp{
		List:      list,
		Total:     total,
		Page:      req.Page,
		Size:      req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
	}
}

// 活动类型中文 → 数字
var activityTypeStrToInt = map[string]int{
	"徒步": 1, "露营": 2, "骑行": 3, "登山": 4, "跑步": 5, "攀岩": 6, "溯溪": 7,
}

// 活动类型数字 → 中文
var activityTypeIntToStr = map[int]string{
	1: "徒步", 2: "露营", 3: "骑行", 4: "登山", 5: "跑步", 6: "攀岩", 7: "溯溪",
}

// 难度中文 → 数字
var difficultyStrToInt = map[string]int{"入门": 1, "中等": 2, "中高等": 3, "高等": 4}

// 难度数字 → 中文
var difficultyIntToStr = map[int]string{1: "入门", 2: "中等", 3: "中高等", 4: "高等"}

// 活动状态中文 → 数字
var activityStatusStrToInt = map[string]int{"报名中": 1, "进行中": 2, "已结束": 3}

// 活动状态数字 → 中文
var activityStatusIntToStr = map[int]string{1: "报名中", 2: "进行中", 3: "已结束"}
