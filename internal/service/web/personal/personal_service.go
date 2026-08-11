package personal

import (
	"context"
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
	"tuyou/pkg/errors"

	"gorm.io/gorm"
)

// 个人主页兜底默认值（数据库里没填时返回给前端）
const (
	defaultAvatar = "https://tuyou1.oss-cn-beijing.aliyuncs.com/images/2026/08/10/%E9%BB%98%E8%AE%A4%E5%A4%B4%E5%83%8F.png" // 头像：小图
	defaultBio    = "这个人很懒，什么都没留下"
	defaultBgImg  = "https://tuyou1.oss-cn-beijing.aliyuncs.com/images/2026/08/10/%E9%BB%98%E8%AE%A4%E8%83%8C%E6%99%AF%E5%9B%BE.jpg" // 背景图：大图
)

type personalService struct {
	db           *gorm.DB
	repo         repoPersonal.PersonalRepository
	noteRepo     repoNote.NoteRepository
	guideRepo    repoGuide.GuideRepository
	activityRepo repoActivity.ActivityRepository
}

// NewPersonalService 创建服务实例
func NewPersonalService(db *gorm.DB, repo repoPersonal.PersonalRepository, noteRepo repoNote.NoteRepository, guideRepo repoGuide.GuideRepository, activityRepo repoActivity.ActivityRepository) PersonalService {
	return &personalService{db: db, repo: repo, noteRepo: noteRepo, guideRepo: guideRepo, activityRepo: activityRepo}
}

// strOr 空串返回默认值，非空返回原值
func strOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// GetProfile 获取个人主页信息
func (s *personalService) GetProfile(memberID uint) (*responsePersonal.Profile, error) {
	member, err := s.repo.FindByID(memberID)
	if err != nil {
		return nil, err
	}

	return &responsePersonal.Profile{
		ID:           strconv.FormatUint(uint64(member.ID), 10),
		Nickname:     strOr(member.Nickname, "用户"+strconv.FormatUint(uint64(member.ID), 10)),
		Avatar:       strOr(member.Avatar, defaultAvatar),
		Bio:          strOr(member.Bio, defaultBio),
		BgImg:        strOr(member.BgImg, defaultBgImg),
		PublishCount: member.PublishCount,
		FollowCount:  member.FollowCount,
		FansCount:    member.FansCount,
	}, nil
}

func (s *personalService) UpdateProfile(memberID uint, req *requestPersonal.UpdateProfileRequest) (*responsePersonal.Profile, error) {
	updates := map[string]interface{}{
		"avatar":   req.Avatar,
		"nickname": req.Nickname,
		"bio":      req.Bio,
		"bg_img":   req.BgImg,
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
			Nickname:   m.Nickname,
			Avatar:     m.Avatar,
			Bio:        m.Bio,
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
			Nickname:   m.Nickname,
			Avatar:     m.Avatar,
			Bio:        m.Bio,
			IsFollowed: followed[m.ID],
		})
	}
	return result, total, nil
}

// ToggleFollow 关注/取关（翻转）。关注/取关与双方计数放在同一个事务里，保证数据一致
func (s *personalService) ToggleFollow(followerID uint, followeeID uint) (*responsePersonal.FollowToggleResponse, error) {
	if followerID == followeeID {
		return nil, errors.New(errors.CodeInvalidParam, "不能关注自己")
	}

	// 校验被关注用户存在（查不到会返回 ErrUserNotFound，顺带挡掉 followeeId=0 的非法请求）
	if _, err := s.repo.FindByID(followeeID); err != nil {
		return nil, err
	}

	var isFollowed bool
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先试取关：删到了 → 说明之前已关注，本次是取关，双方计数 -1
		deleted, err := s.repo.DeleteFollow(tx, followerID, followeeID)
		if err != nil {
			return err
		}
		if deleted {
			if err := s.repo.DecrFollowCount(tx, followerID); err != nil {
				return err
			}
			if err := s.repo.DecrFansCount(tx, followeeID); err != nil {
				return err
			}
			isFollowed = false
			return nil
		}

		// 没删到 → 说明之前未关注，本次是关注，双方计数 +1
		if err := s.repo.CreateFollow(tx, followerID, followeeID); err != nil {
			return err
		}
		if err := s.repo.IncrFollowCount(tx, followerID); err != nil {
			return err
		}
		if err := s.repo.IncrFansCount(tx, followeeID); err != nil {
			return err
		}
		isFollowed = true
		return nil
	})
	if err != nil {
		return nil, err
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

// GetPublishActivities 当前用户发布的活动列表（分页）
func (s *personalService) GetPublishActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error) {
	activities, total, err := s.repo.FindPublishActivities(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(activities) == 0 {
		return []responseActivity.ActivityListItem{}, total, nil
	}
	list, err := s.buildActivityCards(activities, memberID, false)
	if err != nil {
		return nil, total, err
	}
	return list, total, nil
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
	list, err := s.buildActivityCards(activities, memberID, true)
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
				Nickname: m.Nickname,
				Avatar:   m.Avatar,
			}
		}
		cards = append(cards, responsePersonal.NoteCard{
			ID:        strconv.FormatUint(uint64(n.ID), 10),
			Type:      "note",
			Author:    author,
			CoverImg:  cover,
			Title:     n.Title,
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
				Nickname: m.Nickname,
				Avatar:   m.Avatar,
			}
		}
		cards = append(cards, responsePersonal.StrategyCard{
			ID:        strconv.FormatUint(uint64(g.ID), 10),
			Type:      "strategy",
			Author:    author,
			CoverImg:  cover,
			Title:     g.Title,
			Tag:       tag,
			LikeCount: g.LikeCount,
			IsLiked:   likedMap[g.ID],
		})
	}
	return cards, nil
}

// buildActivityCards 把活动实体组装成列表项（批量查领队 + 报名状态）
// joined：true 表示该列表中的活动全部视为已报名（如"我参加的活动"），false 表示全部未报名（如"我发布的活动"）
func (s *personalService) buildActivityCards(activities []entity.Activity, memberID uint, joined bool) ([]responseActivity.ActivityListItem, error) {
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
				Nickname: m.Nickname,
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
			Title:          a.Title,
			CoverImg:       a.CoverImg,
			Category:       activityCategoryIntToStr[a.ActivityType],
			Address:        a.Address,
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
