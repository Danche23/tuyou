package personal

import "tuyou/internal/model/entity"

// PersonalRepository 个人中心仓储接口
type PersonalRepository interface {
	FindFollowedMembers(followerID uint, page, size int) ([]entity.Member, int64, error) // FindFollowedMembers 查询当前用户关注的所有用户信息
	FindFansMembers(memberID uint, page, size int) ([]entity.Member, int64, error)       // FindFansMembers 查询当前用户的粉丝列表
	FindFollowedIDs(followerID uint) (map[uint]bool, error)                              // FindFollowedIDs 查询当前用户关注的所有用户 ID 集合
	CreateFollow(followerID uint, followeeID uint) error                                 // 关注
	DeleteFollow(followerID uint, followeeID uint) (bool, error)                         // 取关
	FindNoteLists(memberID uint, page, size int) ([]entity.Note, int64, error)           // 笔记列表
}
