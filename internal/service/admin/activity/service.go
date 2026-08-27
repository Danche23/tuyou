package activity

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/activity"
	response "tuyou/internal/model/dto/response/admin/activity"
	"tuyou/internal/model/entity"
	repo "tuyou/internal/repository/admin/activity"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/errors"
)

type activityService struct {
	activityRepo repo.ActivityRepository
	webRepo      webActivityWriter
	db           *gorm.DB
	notifier     *serviceMsg.Notifier
}

// NewActivityService 创建后台活动服务实例
// webRepo 复用前台活动仓库的通用写操作（创建活动），避免同表操作重复实现
func NewActivityService(activityRepo repo.ActivityRepository, webRepo webActivityWriter, db *gorm.DB, notifier *serviceMsg.Notifier) ActivityService {
	return &activityService{activityRepo: activityRepo, webRepo: webRepo, db: db, notifier: notifier}
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
		case "收费":
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
	list, err := s.buildListItems(activities)
	if err != nil {
		return nil, err
	}

	return s.buildResp(list, total, req), nil
}

// CreateActivity 创建活动（后台代领队发布）
func (s *activityService) CreateActivity(ctx context.Context, req *request.CreateActivityReq) error {
	// ========== 1. 校验领队 ==========
	leader, err := s.activityRepo.FindLeaderByName(ctx, req.LeaderName)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "查询领队信息失败", err)
	}
	if leader == nil {
		return errors.New(errors.CodeInvalidParam, "领队不存在")
	}

	// ========== 2. 活动类型/难度校验 ==========
	activityType, ok := activityTypeStrToInt[req.ActivityType]
	if !ok {
		return errors.New(errors.CodeInvalidParam, "活动类型不合法")
	}
	difficulty, ok := difficultyStrToInt[req.Difficulty]
	if !ok {
		return errors.New(errors.CodeInvalidParam, "难度不合法")
	}

	// ========== 微信群二维码校验 ==========
	if req.WechatQR == "" {
		return errors.New(errors.CodeInvalidParam, "请上传微信群二维码")
	}

	// ========== 3. 时间解析（yyyy-MM-dd HH:mm，不含秒，与前台一致）==========
	startTime, err := time.ParseInLocation("2006-01-02 15:04", req.StartTime, time.Local)
	if err != nil {
		return errors.New(errors.CodeInvalidParam, "活动开始时间格式不正确")
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
	if err != nil {
		return errors.New(errors.CodeInvalidParam, "活动结束时间格式不正确")
	}
	if !endTime.After(startTime) {
		return errors.New(errors.CodeInvalidParam, "结束日期必须晚于开始日期")
	}

	// ========== 4. 报名时间自动计算 ==========
	now := time.Now()
	signUpStart := now
	signUpEnd := startTime.Add(-48 * time.Hour)

	// ========== 5. 事务写入（建活动）==========
	activity := &entity.Activity{
		LeaderMemberID:  leader.ID,
		Title:           req.Title,
		CoverImg:        req.CoverImg,
		ActivityContent: req.ActivityContent,
		Address:         req.Address,
		StartTime:       startTime,
		EndTime:         endTime,
		SignUpStart:     signUpStart,
		SignUpEnd:       signUpEnd,
		MaxPeople:       req.MaxPeople,
		Price:           req.Price,
		ActivityStatus:  1,
		Difficulty:      difficulty,
		ActivityType:    activityType,
		WechatQR:        req.WechatQR,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		return s.webRepo.Create(tx, activity)
	})
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "创建活动失败", err)
	}

	return nil
}

// GetActivityDetail 获取活动详情（复用 buildListItems 的字段映射，仅追加活动详情内容）
func (s *activityService) GetActivityDetail(ctx context.Context, activityID uint) (*response.ActivityDetailResp, error) {
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取活动详情失败", err)
	}
	if activity == nil {
		return nil, errors.New(errors.CodeNotFound, "活动不存在")
	}

	list, err := s.buildListItems([]entity.Activity{*activity})
	if err != nil {
		return nil, err
	}

	return &response.ActivityDetailResp{
		ActivityListItem: list[0],
		ActivityContent:  activity.ActivityContent,
	}, nil
}

// UpdateActivity 编辑活动（分类/轮播/状态三字段整体覆盖），返回更新后的详情
func (s *activityService) UpdateActivity(ctx context.Context, activityID uint, req *request.UpdateActivityReq) (*response.ActivityDetailResp, error) {
	// 1. 校验活动存在
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取活动失败", err)
	}
	if activity == nil {
		return nil, errors.New(errors.CodeNotFound, "活动不存在")
	}

	// 2. 解析三个必传字段
	activityType, ok := activityTypeStrToInt[req.Category]
	if !ok {
		return nil, errors.New(errors.CodeInvalidParam, "活动类型不合法")
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

	var isDel int
	switch req.Status {
	case "下架":
		isDel = 1
	case "正常":
		isDel = 0
	default:
		return nil, errors.New(errors.CodeInvalidParam, "状态参数不合法")
	}

	// 3. 事务内：更新活动 + 状态变化时同步发布数
	err = s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"activity_type": activityType,
			"is_slide":      isSlide,
			"is_del":        isDel,
		}
		// 下架统一关联 delete_time（前后台一致）：下架写删除时间，恢复正常清除
		if activity.IsDel != isDel {
			if isDel == 1 {
				updates["delete_time"] = time.Now()
			} else {
				updates["delete_time"] = nil
			}
		}
		return s.activityRepo.Update(tx, activityID, updates)
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "编辑活动失败", err)
	}

	// 4. 返回更新后的详情
	return s.GetActivityDetail(ctx, activityID)
}

// DeleteActivity 删除活动（硬删，连带报名记录）
func (s *activityService) DeleteActivity(ctx context.Context, activityID uint) error {
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取活动失败", err)
	}
	if activity == nil {
		return errors.New(errors.CodeNotFound, "活动不存在")
	}

	// 删除前查报名者（删除会连带删报名记录，必须先查）
	memberIDs, err := s.activityRepo.FindJoinMemberIDs(ctx, activityID)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "查询报名者失败", err)
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.activityRepo.Delete(tx, activityID)
	}); err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "删除活动失败", err)
	}

	// 通知所有已报名成员（模板11）
	s.notifier.NotifyActivityDeleted(memberIDs, activity.LeaderMemberID, activity.Title)
	return nil
}

// BatchDeleteActivities 批量硬删除活动（连带报名记录；不存在的 ID 自动跳过）
func (s *activityService) BatchDeleteActivities(ctx context.Context, req *request.BatchDeleteReq) error {
	ids, err := parseActivityIDs(req.IDs)
	if err != nil {
		return errors.New(errors.CodeInvalidParam, "ID参数不合法")
	}

	// 一次查出存在的活动（IN 子句自动去重，重复/不存在的 ID 自动跳过）
	activities, err := s.activityRepo.FindActivitiesByIDs(ctx, ids)
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "获取活动失败", err)
	}
	if len(activities) == 0 {
		return nil
	}

	// 删除前逐个查报名者（删除会连带删报名记录，必须先查）
	joinMemberMap := make(map[uint][]uint, len(activities))
	for i := range activities {
		a := &activities[i]
		memberIDs, err := s.activityRepo.FindJoinMemberIDs(ctx, a.ID)
		if err != nil {
			return errors.NewWithErr(errors.CodeInternalError, "查询报名者失败", err)
		}
		joinMemberMap[a.ID] = memberIDs
	}

	// 单事务逐条硬删
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for i := range activities {
			if err := s.activityRepo.Delete(tx, activities[i].ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return errors.NewWithErr(errors.CodeInternalError, "批量删除活动失败", err)
	}

	// 通知所有已报名成员（模板11）
	for i := range activities {
		a := &activities[i]
		s.notifier.NotifyActivityDeleted(joinMemberMap[a.ID], a.LeaderMemberID, a.Title)
	}
	return nil
}

// ExportActivities 按 ID 批量导出活动列表项（不含软删过滤，已下架也导出）
func (s *activityService) ExportActivities(ctx context.Context, req *request.ExportActivityReq) ([]response.ActivityListItem, error) {
	ids, err := parseActivityIDs(req.IDs)
	if err != nil {
		return nil, errors.New(errors.CodeInvalidParam, "ID参数不合法")
	}

	activities, err := s.activityRepo.FindActivitiesByIDs(ctx, ids)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取活动失败", err)
	}

	return s.buildListItems(activities)
}

// parseActivityIDs 解析活动 ID 列表（兼容 "A0001" 展示格式与纯数字）
func parseActivityIDs(ids []string) ([]uint, error) {
	result := make([]uint, 0, len(ids))
	for _, s := range ids {
		id, err := strconv.ParseUint(strings.TrimPrefix(s, "A"), 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, uint(id))
	}
	return result, nil
}

// ListLeaderNames 获取领队名称列表
func (s *activityService) ListLeaderNames(ctx context.Context) ([]response.LeaderNameItem, error) {
	leaders, err := s.activityRepo.FindLeaders(ctx)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取领队名称列表失败", err)
	}

	list := make([]response.LeaderNameItem, 0, len(leaders))
	for _, l := range leaders {
		list = append(list, response.LeaderNameItem{
			ID:         l.ID,
			LeaderName: l.Nickname,
		})
	}
	return list, nil
}

// buildListItems 批量查询领队并组装列表项
func (s *activityService) buildListItems(activities []entity.Activity) ([]response.ActivityListItem, error) {
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
	// 复用前台活动仓库的批量成员查询（同一张 member 表，不重复实现）
	members, err := s.webRepo.GetMembersByIDs(leaderIDs)
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

		deleteTime := ""
		if a.DeletedAt.Valid {
			deleteTime = a.DeletedAt.Time.Format("2006-01-02 15:04:05")
		}

		list = append(list, response.ActivityListItem{
			ID:             fmt.Sprintf("A%04d", a.ID),
			CoverImg:       a.CoverImg,
			Title:          a.Title,
			LeaderMemberID: strconv.FormatUint(uint64(a.LeaderMemberID), 10),
			LeaderName:     leaderName,
			ActivityType:   activityTypeIntToStr[a.ActivityType],
			StartTime:      a.StartTime.Format("2006-01-02 15:04"),
			EndTime:        a.EndTime.Format("2006-01-02 15:04"),
			Price:          a.Price,
			CurrentPeople:  a.CurrentPeople,
			MaxPeople:      a.MaxPeople,
			Difficulty:     difficultyIntToStr[a.Difficulty],
			IsSlide:        isSlide,
			ActivityStatus: activityStatus,
			Status:         isDel,
			DeleteTime:     deleteTime,
			IsSensitive:    a.IsSensitive,
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
