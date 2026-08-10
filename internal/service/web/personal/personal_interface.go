package personal

import (
	responseActivity "tuyou/internal/model/dto/response/web/activity"
	responseGuide "tuyou/internal/model/dto/response/web/guide"
	responseNote "tuyou/internal/model/dto/response/web/note"
	responsePersonal "tuyou/internal/model/dto/response/web/personal"
)

type PersonalService interface {
	GetFollowedMembers(memberID uint, page, size int) ([]responsePersonal.FollowedMember, int64, error)
	GetFansMembers(memberID uint, page, size int) ([]responsePersonal.FansMember, int64, error)
	ToggleFollow(followerID uint, followeeID uint) (*responsePersonal.FollowToggleResponse, error)
	GetMyNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error)             // 笔记列表
	GetLikedStrategys(memberID uint, page, size int) ([]responseGuide.StrategyCard, int64, error) // 喜欢的攻略

	// 收藏的笔记 / 收藏的攻略 / 点赞的笔记 / 我发布的攻略
	GetFavoriteNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error)
	GetFavoriteStrategys(memberID uint, page, size int) ([]responseGuide.StrategyCard, int64, error)
	GetLikedNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error)
	GetMyStrategys(memberID uint, page, size int) ([]responseGuide.StrategyCard, int64, error)

	// 我发布的活动 / 我参加的活动
	GetPublishActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error)
	GetJoinActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error)
}
