package personal

import (
	"strconv"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
	responseGuide "tuyou/internal/model/dto/response/web/guide"
	responseNote "tuyou/internal/model/dto/response/web/note"
	responsePersonal "tuyou/internal/model/dto/response/web/personal"
	repoPersonal "tuyou/internal/repository/web/personal"
	serviceActivity "tuyou/internal/service/web/activity"
	serviceGuide "tuyou/internal/service/web/guide"
	serviceNote "tuyou/internal/service/web/note"
	"tuyou/pkg/errors"
)

type personalService struct {
	repo        repoPersonal.PersonalRepository
	noteSvc     serviceNote.NoteService
	guideSvc    serviceGuide.GuideService
	activitySvc serviceActivity.ActivityService
}

// NewPersonalService 创建服务实例
func NewPersonalService(repo repoPersonal.PersonalRepository, noteSvc serviceNote.NoteService, guideSvc serviceGuide.GuideService, activitySvc serviceActivity.ActivityService) PersonalService {
	return &personalService{repo: repo, noteSvc: noteSvc, guideSvc: guideSvc, activitySvc: activitySvc}
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

// ToggleFollow 关注/取关（翻转）
func (s *personalService) ToggleFollow(followerID uint, followeeID uint) (*responsePersonal.FollowToggleResponse, error) {
	if followerID == followeeID {
		return nil, errors.New(errors.CodeInvalidParam, "不能关注自己")
	}

	// 先试取关：删到了 → 说明之前已关注，本次是取关
	deleted, err := s.repo.DeleteFollow(followerID, followeeID)
	if err != nil {
		return nil, err
	}
	if deleted {
		return &responsePersonal.FollowToggleResponse{IsFollowed: false}, nil
	}

	// 没删到 → 说明之前未关注，本次是关注
	if err := s.repo.CreateFollow(followerID, followeeID); err != nil {
		return nil, err
	}
	return &responsePersonal.FollowToggleResponse{IsFollowed: true}, nil
}

func (s *personalService) GetMyNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error) {
	notes, total, err := s.repo.FindNoteLists(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}

	if len(notes) == 0 {
		return []responseNote.NoteCard{}, total, nil
	}
	cards, err := s.noteSvc.BuildNoteCards(notes, memberID)
	if err != nil {
		return nil, 0, err
	}
	return cards, total, nil
}

func (s *personalService) GetLikedStrategys(memberID uint, page, size int) ([]responseGuide.StrategyCard, int64, error) {
	strategys, total, err := s.guideSvc.FindLikedStrategys(memberID, page, size)
	if err != nil {
		return nil, 0, err
	}
	if len(strategys) == 0 {
		return []responseGuide.StrategyCard{}, total, nil
	}
	cards, err := s.guideSvc.BuildStrategyCards(strategys, memberID)
	if err != nil {
		return nil, total, err
	}
	return cards, total, nil
}

// GetFavoriteNotes 当前用户收藏的笔记列表（分页）
func (s *personalService) GetFavoriteNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error) {
	return s.noteSvc.GetCollectNotes(memberID, page, size)
}

// GetFavoriteStrategys 当前用户收藏的攻略列表（分页）
func (s *personalService) GetFavoriteStrategys(memberID uint, page, size int) ([]responseGuide.StrategyCard, int64, error) {
	return s.guideSvc.GetCollectStrategys(memberID, page, size)
}

// GetLikedNotes 当前用户点赞的笔记列表（分页）
func (s *personalService) GetLikedNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error) {
	return s.noteSvc.GetLikedNotes(memberID, page, size)
}

// GetMyStrategys 当前用户发布的攻略列表（分页）
func (s *personalService) GetMyStrategys(memberID uint, page, size int) ([]responseGuide.StrategyCard, int64, error) {
	return s.guideSvc.GetPublishStrategys(memberID, page, size)
}

// GetPublishActivities 当前用户发布的活动列表（分页）
func (s *personalService) GetPublishActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error) {
	return s.activitySvc.GetPublishActivities(memberID, page, size)
}

// GetJoinActivities 当前用户报名参加的活动列表（分页）
func (s *personalService) GetJoinActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error) {
	return s.activitySvc.GetJoinActivities(memberID, page, size)
}
