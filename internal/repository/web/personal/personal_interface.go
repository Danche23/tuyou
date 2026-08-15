package personal

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// PersonalRepository 个人中心仓储接口
type PersonalRepository interface {
	FindByID(id uint) (*entity.Member, error)                                            // 个人主页信息
	UpdateProfile(memberID uint, updates map[string]interface{}) error                   // 更新个人信息
	FindFollowedMembers(followerID uint, page, size int) ([]entity.Member, int64, error) // FindFollowedMembers 查询当前用户关注的所有用户信息
	FindFansMembers(memberID uint, page, size int) ([]entity.Member, int64, error)       // FindFansMembers 查询当前用户的粉丝列表
	FindFollowedIDs(followerID uint) (map[uint]bool, error)                              // FindFollowedIDs 查询当前用户关注的所有用户 ID 集合
	ToggleFollow(tx *gorm.DB, followerID uint, followeeID uint) (bool, int, error) // 原子翻转关注状态，返回(新状态, 计数增量)（方案C）
	AdjustFollowCount(tx *gorm.DB, memberID uint, delta int) error                 // 我的关注数 +delta（GREATEST 防负）
	AdjustFansCount(tx *gorm.DB, memberID uint, delta int) error                   // 对方粉丝数 +delta（GREATEST 防负）

	// FindNoteLists 分页查询当前用户发布的笔记（仅上架）
	FindNoteLists(memberID uint, page, size int) ([]entity.Note, int64, error)

	// FindLikedStrategys 分页查询当前用户点赞过的攻略，按点赞时间倒序（仅上架）
	FindLikedStrategys(memberID uint, page, size int) ([]entity.Strategy, int64, error)

	// FindCollectStrategys 分页查询当前用户收藏过的攻略，按收藏时间倒序（仅上架）
	FindCollectStrategys(memberID uint, page, size int) ([]entity.Strategy, int64, error)

	// FindPublishStrategys 分页查询当前用户发布的攻略，按发布时间倒序（仅上架）
	FindPublishStrategys(memberID uint, page, size int) ([]entity.Strategy, int64, error)

	// FindCollectNotes 分页查询当前用户收藏过的笔记，按收藏时间倒序（仅上架）
	FindCollectNotes(memberID uint, page, size int) ([]entity.Note, int64, error)

	// FindLikedNotes 分页查询当前用户点赞过的笔记，按点赞时间倒序（仅上架）
	FindLikedNotes(memberID uint, page, size int) ([]entity.Note, int64, error)

	// FindPublishActivities 分页查询当前用户发布的活动（未删除，按开始时间倒序）
	FindPublishActivities(memberID uint, page, size int) ([]entity.Activity, int64, error)

	// FindJoinActivities 分页查询当前用户报名参加的活动（未取消、活动未删除，按报名时间倒序）
	FindJoinActivities(memberID uint, page, size int) ([]entity.Activity, int64, error)

	// FindNoteImagesByIDs 批量查询多个笔记的图片（返回 noteID -> 图片列表，按 sort 升序）
	FindNoteImagesByIDs(noteIDs []uint) (map[uint][]entity.NoteImage, error)

	// FindNoteTagsByIDs 批量查询多个笔记的标签（返回 noteID -> 标签列表）
	FindNoteTagsByIDs(noteIDs []uint) (map[uint][]entity.Tag, error)

	// FindStrategyImagesByIDs 批量查询多个攻略的图片（返回 strategyID -> 图片列表，按 sort 升序）
	FindStrategyImagesByIDs(strategyIDs []uint) (map[uint][]entity.StrategyImage, error)

	// FindStrategyTagsByIDs 批量查询多个攻略的标签（返回 strategyID -> 标签列表）
	FindStrategyTagsByIDs(strategyIDs []uint) (map[uint][]entity.Tag, error)

	// FindMembersByIDs 批量查询会员信息（已删除用户会被过滤）
	FindMembersByIDs(memberIDs []uint) ([]entity.Member, error)

	// FindLikedNoteIDs 批量查询用户点赞过的笔记ID集合
	FindLikedNoteIDs(noteIDs []uint, memberID uint) (map[uint]bool, error)

	// FindLikedStrategyIDs 批量查询用户点赞过的攻略ID集合
	FindLikedStrategyIDs(strategyIDs []uint, memberID uint) (map[uint]bool, error)
}
