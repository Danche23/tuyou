package personal

import (
	"context"
	stdErrors "errors"
	"strconv"
	"time"
	requestPersonal "tuyou/internal/model/dto/request/web/personal"

	responseActivity "tuyou/internal/model/dto/response/web/activity"
	responsePersonal "tuyou/internal/model/dto/response/web/personal"
	"tuyou/internal/model/entity"
	repoActivity "tuyou/internal/repository/web/activity"
	repoGuide "tuyou/internal/repository/web/guide"
	repoNote "tuyou/internal/repository/web/note"
	repoPersonal "tuyou/internal/repository/web/personal"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/errors"
	"tuyou/pkg/sensitive"

	"gorm.io/gorm"
)

// 个人主页兜底默认值（数据库里没填时返回给前端；默认值本体现在在 entity.Member，注册时已写入库）

type personalService struct {
	db           *gorm.DB
	repo         repoPersonal.PersonalRepository
	noteRepo     repoNote.NoteRepository
	guideRepo    repoGuide.GuideRepository
	activityRepo repoActivity.ActivityRepository
	sensitive    *sensitive.SensitiveMatcher
	notifier     *serviceMsg.Notifier
}

// NewPersonalService 创建服务实例
func NewPersonalService(db *gorm.DB, repo repoPersonal.PersonalRepository, noteRepo repoNote.NoteRepository, guideRepo repoGuide.GuideRepository, activityRepo repoActivity.ActivityRepository, matcher *sensitive.SensitiveMatcher, notifier *serviceMsg.Notifier) PersonalService {
	return &personalService{db: db, repo: repo, noteRepo: noteRepo, guideRepo: guideRepo, activityRepo: activityRepo, sensitive: matcher, notifier: notifier}
}

// strOr 空串返回默认值，非空返回原值
func strOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// findMemberActive 查会员并校验可用（P8）：不存在 → 用户不存在；已禁用 → 账号不可用。
func (s *personalService) findMemberActive(memberID uint) (*entity.Member, error) {
	member, err := s.repo.FindByID(memberID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	if member.Status != 1 {
		return nil, errors.New(errors.CodeForbidden, "账号已被禁用")
	}
	return member, nil
}

// GetProfile 获取个人主页信息
func (s *personalService) GetProfile(memberID uint) (*responsePersonal.Profile, error) {
	member, err := s.repo.FindByID(memberID)
	if err != nil {
		return nil, err
	}

	return &responsePersonal.Profile{
		ID:           strconv.FormatUint(uint64(member.ID), 10),
		Nickname:     strOr(s.mask(member.Nickname), "用户"+strconv.FormatUint(uint64(member.ID), 10)),
		Avatar:       strOr(member.Avatar, entity.DefaultAvatar),
		Bio:          strOr(s.mask(member.Bio), entity.DefaultBio),
		BgImg:        strOr(member.BgImg, entity.DefaultBgImg),
		PublishCount: member.PublishCount,
		FollowCount:  member.FollowCount,
		FansCount:    member.FansCount,
	}, nil
}

// GetMemberCard 他人公开资料卡（登录即可查任意用户；不返回 phone/password；已注销/禁用仍返回资料卡）
func (s *personalService) GetMemberCard(currentMemberID, targetID uint) (*responsePersonal.MemberCard, error) {
	member, err := s.repo.FindByID(targetID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	// 我关注了 TA？ TA 关注了我？（互关判断）
	myFollowed, err := s.repo.FindFollowedIDs(currentMemberID)
	if err != nil {
		return nil, err
	}
	targetFollowed, err := s.repo.FindFollowedIDs(targetID)
	if err != nil {
		return nil, err
	}

	roleStr := "普通会员"
	if member.Role == 0 {
		roleStr = "领队"
	}
	leadingCount := int64(0)
	if member.LeadingCount != nil {
		leadingCount = *member.LeadingCount
	}

	return &responsePersonal.MemberCard{
		ID:            strconv.FormatUint(uint64(member.ID), 10),
		Nickname:      strOr(s.mask(member.Nickname), "用户"+strconv.FormatUint(uint64(member.ID), 10)),
		Avatar:        strOr(member.Avatar, entity.DefaultAvatar),
		Bio:           strOr(s.mask(member.Bio), entity.DefaultBio),
		BgImg:         strOr(member.BgImg, entity.DefaultBgImg),
		Role:          roleStr,
		Status:        member.Status,
		FollowCount:   member.FollowCount,
		FansCount:     member.FansCount,
		PublishCount:  member.PublishCount,
		FavoriteCount: member.FavoriteCount,
		LeadingCount:  leadingCount,
		IsFollowed:    myFollowed[targetID],
		IsFollower:    targetFollowed[currentMemberID],
	}, nil
}

func (s *personalService) UpdateProfile(memberID uint, req *requestPersonal.UpdateProfileRequest) (*responsePersonal.Profile, error) {
	// P8：账号被禁用后不允许写操作
	if _, err := s.findMemberActive(memberID); err != nil {
		return nil, err
	}

	// P6：只写本次传了的字段，避免 strOr 默认值覆盖存量
	updates := map[string]interface{}{}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.BgImg != "" {
		updates["bg_img"] = req.BgImg
	}
	if len(updates) == 0 {
		return nil, errors.New(errors.CodeInvalidParam, "没有要更新的内容")
	}

	if err := s.repo.UpdateProfile(memberID, updates); err != nil {
		return nil, err
	}
	return s.GetProfile(memberID)
}

// GetFollowedMembers 获取我关注的用户列表
func (s *personalService) GetFollowedMembers(memberID uint, page, size int) ([]responsePersonal.FollowedMember, int64, error) {
	members, total, err := s.repo.FindFollowedMembers(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}

	result := make([]responsePersonal.FollowedMember, 0, len(members))
	for _, m := range members {
		result = append(result, responsePersonal.FollowedMember{
			FolloweeID: strconv.FormatUint(uint64(m.ID), 10),
			Nickname:   s.mask(m.Nickname),
			Avatar:     m.Avatar,
			Bio:        s.mask(m.Bio),
			IsFollowed: true, // 关注列表里的人都是我已经关注的
		})
	}
	return result, total, nil
}

// GetFansMembers 获取我的粉丝列表
func (s *personalService) GetFansMembers(memberID uint, page, size int) ([]responsePersonal.FansMember, int64, error) {
	members, total, err := s.repo.FindFansMembers(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}

	// 我关注的用户 ID 集合，用于判断该粉丝是否已与我互关
	followed, err := s.repo.FindFollowedIDs(memberID)
	if err != nil {
		return nil, total, err
	}

	result := make([]responsePersonal.FansMember, 0, len(members))
	for _, m := range members {
		result = append(result, responsePersonal.FansMember{
			FollowerID: strconv.FormatUint(uint64(m.ID), 10),
			Nickname:   s.mask(m.Nickname),
			Avatar:     m.Avatar,
			Bio:        s.mask(m.Bio),
			IsFollowed: followed[m.ID],
		})
	}
	return result, total, nil
}

// ToggleFollow 关注/取关（翻转）：原子 upsert 位翻转 + 双方计数同事务调整。
// 并发安全：唯一索引 idx_follower_followee 作串行器，两个并发请求由 InnoDB 排队，
// 后到者走 ON DUPLICATE KEY UPDATE 翻转，不会出现"删0行误走关注分支"或撞索引 500（方案 C）。
func (s *personalService) ToggleFollow(followerID uint, followeeID uint) (*responsePersonal.FollowToggleResponse, error) {
	if followerID == followeeID {
		return nil, errors.New(errors.CodeInvalidParam, "不能关注自己")
	}

	// P8：当前账号被禁用后不允许写操作（登录态不自动失效，这里兜底）
	if _, err := s.findMemberActive(followerID); err != nil {
		return nil, err
	}

	// P3：被关注用户必须存在（裸 gorm 错误会兜成 500，这里显式转"用户不存在"）
	if _, err := s.repo.FindByID(followeeID); err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	var isFollowed bool
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 原子翻转：不存在→插 status=1；存在→status=1-status。返回(新状态, 计数增量)
		newStatus, delta, err := s.repo.ToggleFollow(tx, followerID, followeeID)
		if err != nil {
			return err
		}
		isFollowed = newStatus
		// 双方计数各加 delta（GREATEST 防负）
		if err := s.repo.AdjustFollowCount(tx, followerID, delta); err != nil {
			return err
		}
		if err := s.repo.AdjustFansCount(tx, followeeID, delta); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 关注成功 → 给被关注者发实时通知（取关不通知）
	if isFollowed {
		s.notifier.NotifyFollow(followeeID, followerID)
	}

	return &responsePersonal.FollowToggleResponse{IsFollowed: isFollowed}, nil
}

// GetMyNotes 当前用户发布的笔记列表（分页）
func (s *personalService) GetMyNotes(memberID uint, page, size int) ([]responsePersonal.NoteCard, int64, error) {
	notes, total, err := s.repo.FindNoteLists(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(notes) == 0 {
		return []responsePersonal.NoteCard{}, total, nil
	}
	cards, err := s.buildNoteCards(notes, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetLikedStrategys 当前用户点赞的攻略列表（分页）
func (s *personalService) GetLikedStrategys(memberID uint, page, size int) ([]responsePersonal.StrategyCard, int64, error) {
	strategys, total, err := s.repo.FindLikedStrategys(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(strategys) == 0 {
		return []responsePersonal.StrategyCard{}, total, nil
	}
	cards, err := s.buildStrategyCards(strategys, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetFavoriteNotes 当前用户收藏的笔记列表（分页）
func (s *personalService) GetFavoriteNotes(memberID uint, page, size int) ([]responsePersonal.NoteCard, int64, error) {
	notes, total, err := s.repo.FindCollectNotes(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(notes) == 0 {
		return []responsePersonal.NoteCard{}, total, nil
	}
	cards, err := s.buildNoteCards(notes, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetFavoriteStrategys 当前用户收藏的攻略列表（分页）
func (s *personalService) GetFavoriteStrategys(memberID uint, page, size int) ([]responsePersonal.StrategyCard, int64, error) {
	strategys, total, err := s.repo.FindCollectStrategys(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(strategys) == 0 {
		return []responsePersonal.StrategyCard{}, total, nil
	}
	cards, err := s.buildStrategyCards(strategys, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetLikedNotes 当前用户点赞的笔记列表（分页）
func (s *personalService) GetLikedNotes(memberID uint, page, size int) ([]responsePersonal.NoteCard, int64, error) {
	notes, total, err := s.repo.FindLikedNotes(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(notes) == 0 {
		return []responsePersonal.NoteCard{}, total, nil
	}
	cards, err := s.buildNoteCards(notes, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetMyStrategys 当前用户发布的攻略列表（分页）
func (s *personalService) GetMyStrategys(memberID uint, page, size int) ([]responsePersonal.StrategyCard, int64, error) {
	strategys, total, err := s.repo.FindPublishStrategys(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(strategys) == 0 {
		return []responsePersonal.StrategyCard{}, total, nil
	}
	cards, err := s.buildStrategyCards(strategys, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetPublishActivities 当前用户发布的活动列表（分页；列表项带全量报名人，含手机号，仅领队本人可见）
func (s *personalService) GetPublishActivities(memberID uint, page, size int) ([]responsePersonal.PublishActivityItem, int64, error) {
	activities, total, err := s.repo.FindPublishActivities(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(activities) == 0 {
		return []responsePersonal.PublishActivityItem{}, total, nil
	}

	baseList, err := s.buildActivityCards(activities, false)
	if err != nil {
		return nil, total, err
	}

	// 查询当前用户对这些活动的实际报名状态
	activityIDs := make([]uint, 0, len(activities))
	for _, a := range activities {
		activityIDs = append(activityIDs, a.ID)
	}
	joinedMap, err := s.activityRepo.FindJoinedActivityIDs(activityIDs, memberID)
	if err != nil {
		return nil, total, err
	}

	// 批量查报名人（本接口请求者即发布者 = 领队，手机号安全）
	participantsMap, err := s.repo.FindParticipantsByActivityIDs(activityIDs)
	if err != nil {
		return nil, total, err
	}

	list := make([]responsePersonal.PublishActivityItem, 0, len(baseList))
	for i, item := range baseList {
		item.Joined = joinedMap[activities[i].ID]
		list = append(list, responsePersonal.PublishActivityItem{
			ActivityListItem: item,
			Participants:     s.buildParticipants(participantsMap[activities[i].ID]),
		})
	}
	return list, total, nil
}

// buildParticipants 报名人组装（昵称打星；手机号透传——仅领队可见出口调用）
func (s *personalService) buildParticipants(members []entity.Member) []responsePersonal.ParticipantInfo {
	list := make([]responsePersonal.ParticipantInfo, 0, len(members))
	for _, m := range members {
		list = append(list, responsePersonal.ParticipantInfo{
			ID:       strconv.FormatUint(uint64(m.ID), 10),
			Avatar:   m.Avatar,
			Nickname: s.mask(m.Nickname),
			Phone:    m.Phone,
		})
	}
	return list
}

// GetJoinActivities 当前用户报名参加的活动列表（分页）
func (s *personalService) GetJoinActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error) {
	activities, total, err := s.repo.FindJoinActivities(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(activities) == 0 {
		return []responseActivity.ActivityListItem{}, total, nil
	}
	list, err := s.buildActivityCards(activities, true)
	if err != nil {
		return nil, total, err
	}
	return list, total, nil
}

// ========== 组装 ==========

// buildNoteCards 把笔记实体组装成卡片（批量查询图片/标签/作者/点赞状态，避免 N+1）
func (s *personalService) buildNoteCards(notes []entity.Note, memberID uint) ([]responsePersonal.NoteCard, error) {
	if len(notes) == 0 {
		return []responsePersonal.NoteCard{}, nil
	}

	noteIDs := make([]uint, 0, len(notes))
	authorIDSet := make(map[uint]bool, len(notes))
	authorIDs := make([]uint, 0, len(notes))
	for _, n := range notes {
		noteIDs = append(noteIDs, n.ID)
		if !authorIDSet[n.MemberID] {
			authorIDSet[n.MemberID] = true
			authorIDs = append(authorIDs, n.MemberID)
		}
	}

	// 批量查询（走个人模块自己的仓储，避免改 note 仓储）
	imagesMap, err := s.repo.FindNoteImagesByIDs(noteIDs)
	if err != nil {
		return nil, err
	}
	tagsMap, err := s.repo.FindNoteTagsByIDs(noteIDs)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.FindMembersByIDs(authorIDs)
	if err != nil {
		return nil, err
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}
	likedMap := s.readNoteLikedStatuses(noteIDs, memberID)

	cards := make([]responsePersonal.NoteCard, 0, len(notes))
	for _, n := range notes {
		cover := ""
		if images := imagesMap[n.ID]; len(images) > 0 {
			cover = images[0].URL
		}
		tag := ""
		if tags := tagsMap[n.ID]; len(tags) > 0 {
			tag = tags[0].TagName
		}
		var author *responsePersonal.AuthorInfo
		if m, ok := memberMap[n.MemberID]; ok {
			author = &responsePersonal.AuthorInfo{
				ID:       strconv.FormatUint(uint64(m.ID), 10),
				Nickname: s.mask(m.Nickname),
				Avatar:   m.Avatar,
			}
		}
		cards = append(cards, responsePersonal.NoteCard{
			ID:        strconv.FormatUint(uint64(n.ID), 10),
			Type:      "note",
			Author:    author,
			CoverImg:  cover,
			IsVideo:   n.VideoURL != "",
			Title:     s.mask(n.Title),
			Tag:       tag,
			LikeCount: n.LikeCount,
			IsLiked:   likedMap[n.ID],
		})
	}
	return cards, nil
}

// buildStrategyCards 把攻略实体组装成卡片（批量查询图片/标签/作者/点赞状态，避免 N+1）
func (s *personalService) buildStrategyCards(strategies []entity.Strategy, memberID uint) ([]responsePersonal.StrategyCard, error) {
	if len(strategies) == 0 {
		return []responsePersonal.StrategyCard{}, nil
	}

	strategyIDs := make([]uint, 0, len(strategies))
	authorIDSet := make(map[uint]bool, len(strategies))
	authorIDs := make([]uint, 0, len(strategies))
	for _, g := range strategies {
		strategyIDs = append(strategyIDs, g.ID)
		if !authorIDSet[g.MemberID] {
			authorIDSet[g.MemberID] = true
			authorIDs = append(authorIDs, g.MemberID)
		}
	}

	// 批量查询（走个人模块自己的仓储，避免改 guide 仓储）
	imagesMap, err := s.repo.FindStrategyImagesByIDs(strategyIDs)
	if err != nil {
		return nil, err
	}
	tagsMap, err := s.repo.FindStrategyTagsByIDs(strategyIDs)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.FindMembersByIDs(authorIDs)
	if err != nil {
		return nil, err
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}
	likedMap := s.readStrategyLikedStatuses(strategyIDs, memberID)

	cards := make([]responsePersonal.StrategyCard, 0, len(strategies))
	for _, g := range strategies {
		cover := ""
		if images := imagesMap[g.ID]; len(images) > 0 {
			cover = images[0].URL
		}
		tag := ""
		if tags := tagsMap[g.ID]; len(tags) > 0 {
			tag = tags[0].TagName
		}
		var author *responsePersonal.AuthorInfo
		if m, ok := memberMap[g.MemberID]; ok {
			author = &responsePersonal.AuthorInfo{
				ID:       strconv.FormatUint(uint64(m.ID), 10),
				Nickname: s.mask(m.Nickname),
				Avatar:   m.Avatar,
			}
		}
		cards = append(cards, responsePersonal.StrategyCard{
			ID:        strconv.FormatUint(uint64(g.ID), 10),
			Type:      "strategy",
			Author:    author,
			CoverImg:  cover,
			IsVideo:   g.VideoURL != "",
			Title:     s.mask(g.Title),
			Tag:       tag,
			LikeCount: g.LikeCount,
			IsLiked:   likedMap[g.ID],
		})
	}
	return cards, nil
}

// buildActivityCards 把活动实体组装成列表项（批量查领队 + 报名状态）
// joined：true 表示该列表中的活动全部视为已报名（如"我参加的活动"），false 表示全部未报名（如"我发布的活动"）
func (s *personalService) buildActivityCards(activities []entity.Activity, joined bool) ([]responseActivity.ActivityListItem, error) {
	// 批量获取领队信息
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
		return nil, err
	}
	memberMap := make(map[uint]*entity.Member, len(members))
	for i := range members {
		memberMap[members[i].ID] = &members[i]
	}

	// 报名状态
	var joinedMap map[uint]bool
	if joined {
		joinedMap = make(map[uint]bool, len(activities))
		for _, a := range activities {
			joinedMap[a.ID] = true
		}
	}

	return s.buildActivityItems(activities, memberMap, joinedMap)
}

// buildActivityItems 组装活动列表项（领队信息 + 报名状态 + 活动状态按时间计算）
func (s *personalService) buildActivityItems(activities []entity.Activity, memberMap map[uint]*entity.Member, joinedMap map[uint]bool) ([]responseActivity.ActivityListItem, error) {
	list := make([]responseActivity.ActivityListItem, 0, len(activities))
	for _, a := range activities {
		var leader *responseActivity.LeaderInfo
		if m, ok := memberMap[a.LeaderMemberID]; ok {
			leader = &responseActivity.LeaderInfo{
				ID:       strconv.FormatUint(uint64(m.ID), 10),
				Nickname: s.mask(m.Nickname),
				Avatar:   m.Avatar,
			}
		}

		_, statusStr := computeActivityStatus(a.SignUpEnd, a.EndTime)

		joined := false
		if joinedMap != nil {
			joined = joinedMap[a.ID]
		}

		list = append(list, responseActivity.ActivityListItem{
			ID:             strconv.FormatUint(uint64(a.ID), 10),
			Leader:         leader,
			Title:          s.mask(a.Title),
			CoverImg:       a.CoverImg,
			Category:       activityCategoryIntToStr[a.ActivityType],
			Address:        s.mask(a.Address),
			StartTime:      a.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:        a.EndTime.Format("2006-01-02 15:04:05"),
			SignUpEnd:      a.SignUpEnd.Format("2006-01-02 15:04:05"),
			Price:          a.Price,
			MaxPeople:      a.MaxPeople,
			CurrentPeople:  a.CurrentPeople,
			Difficulty:     activityDifficultyIntToStr[a.Difficulty],
			ActivityStatus: statusStr,
			Joined:         joined,
		})
	}
	return list, nil
}

// ========== 点赞状态（Redis 缓存优先 → MySQL 回填） ==========

// readNoteLikedStatuses 批量读取笔记的点赞状态（Redis 缓存优先 → MySQL 批量回填）
func (s *personalService) readNoteLikedStatuses(noteIDs []uint, memberID uint) map[uint]bool {
	ctx := context.Background()
	result := make(map[uint]bool, len(noteIDs))

	uncached := make([]uint, 0, len(noteIDs))
	for _, id := range noteIDs {
		isLiked, found, err := s.noteRepo.GetCachedLikeStatus(ctx, id, memberID)
		if err == nil && found {
			result[id] = isLiked
		} else {
			uncached = append(uncached, id)
		}
	}

	if len(uncached) > 0 {
		likedMap, err := s.repo.FindLikedNoteIDs(uncached, memberID)
		if err != nil {
			// 批量查询失败，未命中的卡片默认未点赞
			for _, id := range uncached {
				result[id] = false
			}
		} else {
			for _, id := range uncached {
				isLiked := likedMap[id]
				result[id] = isLiked
				s.noteRepo.SetCachedLikeStatus(ctx, id, memberID, isLiked)
			}
		}
	}
	return result
}

// readStrategyLikedStatuses 批量读取攻略的点赞状态（Redis 缓存优先 → MySQL 批量回填）
func (s *personalService) readStrategyLikedStatuses(strategyIDs []uint, memberID uint) map[uint]bool {
	ctx := context.Background()
	result := make(map[uint]bool, len(strategyIDs))

	uncached := make([]uint, 0, len(strategyIDs))
	for _, id := range strategyIDs {
		isLiked, found, err := s.guideRepo.GetCachedLikeStatus(ctx, id, memberID)
		if err == nil && found {
			result[id] = isLiked
		} else {
			uncached = append(uncached, id)
		}
	}

	if len(uncached) > 0 {
		likedMap, err := s.repo.FindLikedStrategyIDs(uncached, memberID)
		if err != nil {
			// 批量查询失败，未命中的卡片默认未点赞
			for _, id := range uncached {
				result[id] = false
			}
		} else {
			for _, id := range uncached {
				isLiked := likedMap[id]
				result[id] = isLiked
				s.guideRepo.SetCachedLikeStatus(ctx, id, memberID, isLiked)
			}
		}
	}
	return result
}

// ========== 活动状态工具（与 activity 模块保持一致的展示规则） ==========

// computeActivityStatus 按时间计算活动状态（报名中/进行中/已结束）
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

// activityCategoryIntToStr 活动分类数值 → 名称
var activityCategoryIntToStr = map[int]string{
	1: "徒步", 2: "露营", 3: "骑行", 4: "登山", 5: "跑步", 6: "攀岩", 7: "溯溪",
}

// activityDifficultyIntToStr 活动难度数值 → 名称
var activityDifficultyIntToStr = map[int]string{
	1: "入门", 2: "中等", 3: "中高等", 4: "高等",
}

// mask 前台文本出口统一打星（nil 安全）
func (s *personalService) mask(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}
