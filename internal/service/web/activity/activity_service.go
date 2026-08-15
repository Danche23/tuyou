package activity

import (
	"strconv"
	"time"

	requestActivity "tuyou/internal/model/dto/request/web/activity"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
	"tuyou/internal/model/entity"
	repoActivity "tuyou/internal/repository/web/activity"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"

	"gorm.io/gorm"
)

type activityService struct {
	activityRepo repoActivity.ActivityRepository
	db           *gorm.DB
	notifier     *serviceMsg.Notifier
}

// NewActivityService 创建活动服务实例
func NewActivityService(activityRepo repoActivity.ActivityRepository, db *gorm.DB, notifier *serviceMsg.Notifier) ActivityService {
	return &activityService{activityRepo: activityRepo, db: db, notifier: notifier}
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
		return nil, errors.New(errors.CodeInvalidParam, "活动开始时间格式不正确")
	}
	if !startTime.After(time.Now().Add(48 * time.Hour)) {
		return nil, errors.New(errors.CodeInvalidParam, "活动开始时间必须晚于当前时间48小时")
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
		Leader:          &responseActivity.LeaderInfo{ID: strconv.FormatUint(uint64(author.ID), 10), Nickname: author.Nickname, Avatar: author.Avatar, Bio: author.Bio},
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

	// 当前用户是否已关注领队（自己发布的活动默认不关注）
	isFollowed := false
	if activity.LeaderMemberID != memberID {
		isFollowed, err = s.activityRepo.ExistsFollow(memberID, activity.LeaderMemberID)
		if err != nil {
			return nil, errors.New(errors.CodeInternalError, "查询关注状态失败")
		}
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
		Leader:          &responseActivity.LeaderInfo{ID: strconv.FormatUint(uint64(leader.ID), 10), Nickname: leader.Nickname, Avatar: leader.Avatar, Bio: leader.Bio},
		IsOwner:         activity.LeaderMemberID == memberID,
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
		IsFollowed:      isFollowed,
		Joined:          joined,
		Members:         memberInfos,
	}, nil
}

// ToggleSignup 报名/取消报名
func (s *activityService) ToggleSignup(activityID uint, memberID uint) (*responseActivity.SignupToggleResponse, error) {
	activity, err := s.activityRepo.FindByID(activityID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "活动不存在")
	}

	// 同步活动状态
	realStatusInt, _ := computeActivityStatus(activity.SignUpEnd, activity.EndTime)
	if activity.ActivityStatus != realStatusInt {
		if err := s.activityRepo.UpdateActivityStatus(activityID, realStatusInt); err != nil {
			logger.Errorf("同步活动状态失败 activityID=%d err=%v", activityID, err)
		}
	}

	joined, err := s.activityRepo.ExistsJoin(activityID, memberID)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "查询报名状态失败")
	}

	if joined {
		// ========== 取消报名 ==========
		// 仅"报名中"（状态1）允许取消；进行中/已结束均不可取消（后端兜底，与前端按钮一致）
		if realStatusInt != 1 {
			return nil, errors.New(errors.CodeInvalidParam, "活动已开始，无法取消报名")
		}
		err = s.db.Transaction(func(tx *gorm.DB) error {
			deleted, err := s.activityRepo.DeleteJoin(tx, activityID, memberID)
			if err != nil {
				return err
			}
			if !deleted {
				return nil // 本来就没报名/已被并发取消，不减人数
			}
			return s.activityRepo.DecrCurrentPeople(tx, activityID)
		})
		if err != nil {
			logger.Errorf("取消报名事务失败 activityID=%d memberID=%d err=%v", activityID, memberID, err)
			return nil, errors.NewWithErr(errors.CodeInternalError, "报名操作失败", err)
		}

		// 事务提交后重查真实报名人数（并发下不用旧值±1，保证响应实时准确）
		currentPeople, qErr := s.activityRepo.GetCurrentPeople(s.db, activityID)
		if qErr != nil {
			logger.Errorf("查询报名人数失败 activityID=%d err=%v", activityID, qErr)
			currentPeople = activity.CurrentPeople - 1 // 兜底用旧值
		}
		return &responseActivity.SignupToggleResponse{
			IsJoined:      false,
			CurrentPeople: currentPeople,
		}, nil
	}

	// ========== 报名 ==========
	if realStatusInt != 1 {
		return nil, errors.New(errors.CodeInvalidParam, "不在报名时间内")
	}
	var bizErr error
	alreadyJoined := false
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 先插报名记录，用唯一索引兜底检测重复点击（并发双击时第二次返回 false）
		created, cErr := s.activityRepo.CreateJoin(tx, activityID, memberID)
		if cErr != nil {
			return cErr
		}
		if !created {
			// 重复报名：人数不动、不报错，直接视为已报名
			alreadyJoined = true
			return nil
		}
		// 新报名才加人数（WHERE current_people < max_people 原子防超卖）
		affected, incrErr := s.activityRepo.IncrCurrentPeople(tx, activityID, activity.MaxPeople)
		if incrErr != nil {
			return incrErr
		}
		if affected == 0 {
			bizErr = errors.New(errors.CodeInvalidParam, "活动人数已满")
			return bizErr // 回滚，连刚插入的报名记录也撤销
		}
		return nil
	})
	if bizErr != nil {
		return nil, bizErr
	}
	if err != nil {
		logger.Errorf("报名事务失败 activityID=%d memberID=%d err=%v", activityID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "报名操作失败", err)
	}

	// 事务提交后重查真实报名人数（并发下不用旧值±1，保证响应实时准确）
	currentPeople, qErr := s.activityRepo.GetCurrentPeople(s.db, activityID)
	if qErr != nil {
		logger.Errorf("查询报名人数失败 activityID=%d err=%v", activityID, qErr)
		if alreadyJoined {
			currentPeople = activity.CurrentPeople // 兜底用旧值
		} else {
			currentPeople = activity.CurrentPeople + 1 // 兜底用旧值
		}
	}
	if alreadyJoined {
		// 重复点击：返回已报名状态，不发通知、人数不重复加
		return &responseActivity.SignupToggleResponse{
			IsJoined:      true,
			CurrentPeople: currentPeople,
		}, nil
	}

	// 报名成功 → 给活动领队发实时通知（取消报名不通知）
	s.notifier.NotifySignup(activity.LeaderMemberID, memberID, activity.Title)
	// 给报名者本人发"报名成功"系统通知
	s.notifier.NotifySignupSuccess(memberID, activity.Title)

	return &responseActivity.SignupToggleResponse{
		IsJoined:      true,
		CurrentPeople: currentPeople,
	}, nil
}

// DeleteActivity 删除活动（逻辑删除，仅创建者）
func (s *activityService) DeleteActivity(activityID uint, memberID uint) error {
	activity, err := s.activityRepo.FindByID(activityID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "活动不存在")
	}
	if activity.LeaderMemberID != memberID {
		return errors.New(errors.CodeForbidden, "无权删除他人活动")
	}

	// 删除前先取已报名成员，删除成功后用于通知（best-effort，查不到不阻塞删除）
	members, err := s.activityRepo.FindJoinedMembers(activityID)
	if err != nil {
		logger.Errorf("删除前查询报名成员失败 activityID=%d err=%v", activityID, err)
		members = nil
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.activityRepo.SoftDelete(tx, activityID); err != nil {
			return err
		}
		return s.activityRepo.DecrPublishCount(tx, memberID)
	})
	if err != nil {
		logger.Errorf("删除活动事务失败 activityID=%d memberID=%d err=%v", activityID, memberID, err)
		return errors.NewWithErr(errors.CodeInternalError, "删除活动失败", err)
	}

	// 删除成功 → 给所有已报名成员发活动删除通知
	memberIDs := make([]uint, 0, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.ID)
	}
	s.notifier.NotifyActivityDeleted(memberIDs, memberID, activity.Title)

	return nil
}

// GetActivityList 获取活动列表
func (s *activityService) GetActivityList(memberID uint, req *requestActivity.ListActivityRequest) ([]responseActivity.ActivityListItem, int64, error) {
	// ========== 1. 默认分页参数 ==========
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 10
	}

	// ========== 2. 分类校验与转换 ==========
	var category int
	if req.Category != "" {
		var ok bool
		category, ok = categoryStrToInt[req.Category]
		if !ok {
			return nil, 0, errors.New(errors.CodeInvalidParam, "活动类型不合法")
		}
	}

	// ========== 3. 查询活动列表 ==========
	activities, total, err := s.activityRepo.FindList(category, req.Page, req.Size)
	if err != nil {
		return nil, 0, errors.New(errors.CodeInternalError, "获取活动列表失败")
	}

	if len(activities) == 0 {
		return []responseActivity.ActivityListItem{}, total, nil
	}

	// ========== 4. 组装响应：复用 BuildActivityItems（唯一实现，search 模块共用） ==========
	list, err := s.BuildActivityItems(memberID, activities)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// BuildActivityItems 组装活动列表项（唯一实现，GetActivityList 与 search 模块共用）
func (s *activityService) BuildActivityItems(memberID uint, activities []entity.Activity) ([]responseActivity.ActivityListItem, error) {
	// ========== 1. 批量获取领队信息 ==========
	leaderIDSet := make(map[uint]bool)
	uniqueLeaderIDs := make([]uint, 0)
	for _, a := range activities {
		if !leaderIDSet[a.LeaderMemberID] {
			leaderIDSet[a.LeaderMemberID] = true
			uniqueLeaderIDs = append(uniqueLeaderIDs, a.LeaderMemberID)
		}
	}
	members, err := s.activityRepo.GetMembersByIDs(uniqueLeaderIDs)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "查询领队信息失败")
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}

	// ========== 2. 批量查询当前用户的报名状态 ==========
	activityIDs := make([]uint, len(activities))
	for i, a := range activities {
		activityIDs[i] = a.ID
	}
	joinedMap, err := s.activityRepo.FindJoinedActivityIDs(activityIDs, memberID)
	if err != nil {
		return nil, errors.New(errors.CodeInternalError, "查询报名状态失败")
	}

	// ========== 3. 组装响应 ==========
	list := make([]responseActivity.ActivityListItem, 0, len(activities))
	for _, a := range activities {
		// 领队信息
		var leader *responseActivity.LeaderInfo
		if m, ok := memberMap[a.LeaderMemberID]; ok {
			leader = &responseActivity.LeaderInfo{
				ID:       strconv.FormatUint(uint64(m.ID), 10),
				Nickname: m.Nickname,
				Avatar:   m.Avatar,
				Bio:      m.Bio,
			}
		}

		// 活动状态按时间计算
		_, statusStr := computeActivityStatus(a.SignUpEnd, a.EndTime)

		list = append(list, responseActivity.ActivityListItem{
			ID:             strconv.FormatUint(uint64(a.ID), 10),
			Leader:         leader,
			Title:          a.Title,
			CoverImg:       a.CoverImg,
			Category:       categoryIntToStr[a.ActivityType],
			Address:        a.Address,
			StartTime:      a.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:        a.EndTime.Format("2006-01-02 15:04:05"),
			SignUpEnd:      a.SignUpEnd.Format("2006-01-02 15:04:05"),
			Price:          a.Price,
			MaxPeople:      a.MaxPeople,
			CurrentPeople:  a.CurrentPeople,
			Difficulty:     difficultyIntToStr[a.Difficulty],
			ActivityStatus: statusStr,
			Joined:         joinedMap[a.ID],
		})
	}

	return list, nil
}

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
