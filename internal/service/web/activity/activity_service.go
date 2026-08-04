package activity

import (
	"strconv"
	"time"

	requestActivity "tuyou/internal/model/dto/request/web/activity"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
	"tuyou/internal/model/entity"
	repoActivity "tuyou/internal/repository/web/activity"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"

	"gorm.io/gorm"
)

type activityService struct {
	activityRepo repoActivity.ActivityRepository
	db           *gorm.DB
}

// NewActivityService 创建活动服务实例
func NewActivityService(activityRepo repoActivity.ActivityRepository, db *gorm.DB) ActivityService {
	return &activityService{activityRepo: activityRepo, db: db}
}

// CreateActivity 创建活动（仅领队 role=0 可创建）
func (s *activityService) CreateActivity(memberID uint, req *requestActivity.CreateActivityRequest) (*responseActivity.CreateActivityResponse, error) {
	// ========== 校验领队身份 ==========
	author, err := s.activityRepo.GetAuthor(memberID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "用户不存在")
	}
	if author.Role != 0 {
		return nil, errors.New(errors.CodeForbidden, "仅领队可发布活动")
	}

	// ========== 分类校验 ==========
	activityType, ok := categoryStrToInt[req.Category]
	if !ok {
		return nil, errors.New(errors.CodeInvalidParam, "活动类型不合法")
	}

	// ========== 难度校验 ==========
	difficulty, ok := difficultyStrToInt[req.Difficulty]
	if !ok {
		return nil, errors.New(errors.CodeInvalidParam, "难度不合法")
	}

	// ========== 时间解析 ==========
	startTime, err := time.ParseInLocation("2006-01-02 15:04", req.StartTime, time.Local)
	if err != nil {
		return nil, errors.New(errors.CodeInvalidParam, "开始时间格式不正确")
	}
	if !startTime.After(time.Now()) {
		return nil, errors.New(errors.CodeInvalidParam, "开始时间必须晚于当前时间")
	}

	endTime, err := time.ParseInLocation("2006-01-02 15:04", req.EndTime, time.Local)
	if err != nil {
		return nil, errors.New(errors.CodeInvalidParam, "结束时间格式不正确")
	}
	if !endTime.After(startTime) {
		return nil, errors.New(errors.CodeInvalidParam, "结束时间必须晚于开始时间")
	}

	// ========== 报名时间自动计算 ==========
	now := time.Now()
	signUpStart := now
	signUpEnd := startTime.Add(-48 * time.Hour)

	// ========== 事务写入 ==========
	activity := &entity.Activity{
		LeaderMemberID:  memberID,
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
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.activityRepo.Create(tx, activity); err != nil {
			return err
		}
		return s.activityRepo.IncrPublishCount(tx, memberID)
	})
	if err != nil {
		logger.Errorf("创建活动事务失败 memberID=%d err=%v", memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "创建活动失败", err)
	}

	// ========== 组装响应 ==========
	return &responseActivity.CreateActivityResponse{
		ID:              strconv.FormatUint(uint64(activity.ID), 10),
		Leader:          &responseActivity.LeaderInfo{ID: strconv.FormatUint(uint64(author.ID), 10), Nickname: author.Nickname, Avatar: author.Avatar},
		Title:           activity.Title,
		CoverImg:        activity.CoverImg,
		Category:        req.Category,
		Address:         activity.Address,
		ActivityContent: activity.ActivityContent,
		StartTime:       activity.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:         activity.EndTime.Format("2006-01-02 15:04:05"),
		SignUpStart:     signUpStart.Format("2006-01-02 15:04:05"),
		SignUpEnd:       signUpEnd.Format("2006-01-02 15:04:05"),
		MaxPeople:       activity.MaxPeople,
		CurrentPeople:   activity.CurrentPeople,
		Price:           activity.Price,
		Difficulty:      req.Difficulty,
		Joined:          false,
	}, nil
}

// GetActivityDetail 获取活动详情
func (s *activityService) GetActivityDetail(activityID uint, memberID uint) (*responseActivity.ActivityDetailResponse, error) {
	activity, err := s.activityRepo.FindByID(activityID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "活动不存在")
	}

	// 领队信息
	leader, err := s.activityRepo.GetAuthor(activity.LeaderMemberID)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "获取领队信息失败")
	}

	// 已报名成员
	members, err := s.activityRepo.FindJoinedMembers(activityID)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "查询报名成员失败")
	}
	memberInfos := make([]responseActivity.MemberInfo, 0, len(members))
	for _, m := range members {
		memberInfos = append(memberInfos, responseActivity.MemberInfo{
			ID:       strconv.FormatUint(uint64(m.ID), 10),
			Nickname: m.Nickname,
			Avatar:   m.Avatar,
		})
	}

	// 当前用户是否已报名
	joined, err := s.activityRepo.ExistsJoin(activityID, memberID)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "查询报名状态失败")
	}

	// 活动状态按时间自动计算，和 DB 不一致时同步更新
	realStatusInt, activityStatus := computeActivityStatus(activity.SignUpEnd, activity.EndTime)
	if activity.ActivityStatus != realStatusInt {
		if err := s.activityRepo.UpdateActivityStatus(activityID, realStatusInt); err != nil {
			logger.Errorf("同步活动状态失败 activityID=%d old=%d new=%d err=%v", activityID, activity.ActivityStatus, realStatusInt, err)
		}
	}

	return &responseActivity.ActivityDetailResponse{
		ID:              strconv.FormatUint(uint64(activity.ID), 10),
		Leader:          &responseActivity.LeaderInfo{ID: strconv.FormatUint(uint64(leader.ID), 10), Nickname: leader.Nickname, Avatar: leader.Avatar},
		Title:           activity.Title,
		CoverImg:        activity.CoverImg,
		Category:        categoryIntToStr[activity.ActivityType],
		Address:         activity.Address,
		ActivityContent: activity.ActivityContent,
		StartTime:       activity.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:         activity.EndTime.Format("2006-01-02 15:04:05"),
		SignUpStart:     activity.SignUpStart.Format("2006-01-02 15:04:05"),
		SignUpEnd:       activity.SignUpEnd.Format("2006-01-02 15:04:05"),
		MaxPeople:       activity.MaxPeople,
		CurrentPeople:   activity.CurrentPeople,
		Price:           activity.Price,
		Difficulty:      difficultyIntToStr[activity.Difficulty],
		ActivityStatus:  activityStatus,
		Joined:          joined,
		Members:         memberInfos,
	}, nil
}

// ────────── 枚举映射 ──────────

var categoryStrToInt = map[string]int{
	"徒步": 1, "露营": 2, "骑行": 3, "登山": 4, "跑步": 5, "攀岩": 6, "溯溪": 7,
}

var categoryIntToStr = map[int]string{
	1: "徒步", 2: "露营", 3: "骑行", 4: "登山", 5: "跑步", 6: "攀岩", 7: "溯溪",
}

var difficultyStrToInt = map[string]int{
	"入门": 1, "中等": 2, "中高等": 3, "高等": 4,
}

var difficultyIntToStr = map[int]string{
	1: "入门", 2: "中等", 3: "中高等", 4: "高等",
}

func computeActivityStatus(signUpEnd, endTime time.Time) (int, string) {
	now := time.Now()
	if now.After(endTime) {
		return 3, "已结束"
	}
	if !now.Before(signUpEnd) {
		return 2, "进行中"
	}
	return 1, "报名中"
}
