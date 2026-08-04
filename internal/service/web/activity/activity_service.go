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

// 分类中文 → 数据库 int 映射（写死枚举，无独立分类接口）
var categoryStrToInt = map[string]int{
	"徒步": 1, "露营": 2, "骑行": 3, "登山": 4, "跑步": 5, "攀岩": 6, "溯溪": 7,
}

// 难度中文 → 数据库 int 映射
var difficultyStrToInt = map[string]int{
	"入门": 1, "中等": 2, "中高等": 3, "高等": 4,
}

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
