package personal

import (
	requestPersonal "tuyou/internal/model/dto/request/web/personal"
	responseActivity "tuyou/internal/model/dto/response/web/activity"
	responsePersonal "tuyou/internal/model/dto/response/web/personal"
)

type PersonalService interface {
	GetProfile(memberID uint) (*responsePersonal.Profile, error)
	GetMemberCard(currentMemberID, targetID uint) (*responsePersonal.MemberCard, error) // 他人公开资料卡（用户详情）
	UpdateProfile(memberID uint, req *requestPersonal.UpdateProfileRequest) (*responsePersonal.Profile, error) // 编辑个人信息
	GetFollowedMembers(memberID uint, page, size int) ([]responsePersonal.FollowedMember, int64, error)
	GetFansMembers(memberID uint, page, size int) ([]responsePersonal.FansMember, int64, error)
	ToggleFollow(followerID uint, followeeID uint) (*responsePersonal.FollowToggleResponse, error)
	GetMyNotes(memberID uint, page, size int) ([]responsePersonal.NoteCard, int64, error)            // 我的笔记
	GetLikedStrategys(memberID uint, page, size int) ([]responsePersonal.StrategyCard, int64, error) // 点赞的攻略

	// 收藏的笔记 / 收藏的攻略 / 点赞的笔记 / 我发布的攻略
	GetFavoriteNotes(memberID uint, page, size int) ([]responsePersonal.NoteCard, int64, error)
	GetFavoriteStrategys(memberID uint, page, size int) ([]responsePersonal.StrategyCard, int64, error)
	GetLikedNotes(memberID uint, page, size int) ([]responsePersonal.NoteCard, int64, error)
	GetMyStrategys(memberID uint, page, size int) ([]responsePersonal.StrategyCard, int64, error)

	// 我发布的活动 / 我参加的活动
	GetPublishActivities(memberID uint, page, size int) ([]responsePersonal.PublishActivityItem, int64, error) // 我发布的活动（带报名人）
	GetJoinActivities(memberID uint, page, size int) ([]responseActivity.ActivityListItem, int64, error)
}
