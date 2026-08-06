package personal

import "tuyou/internal/model/entity"

// PersonalRepository 个人中心仓储接口
type PersonalRepository interface {
	// FindFollowedMembers 查询当前用户关注的所有用户信息
	// followerID 为当前登录用户的 member.id，返回其关注的人（followee）的 member 记录
	// 注意：查询仅 Select 了 nickname/avatar/bio/member_type 四列，entity.Member 其余字段为零值
	FindFollowedMembers(followerID uint) ([]entity.Member, error)
	// FindFansMembers 查询当前用户的粉丝列表
	// memberID 为当前登录用户的 member.id，返回关注了他的人（follower）的 member 记录
	// 注意：查询仅 Select 了 nickname/avatar/bio/member_type 四列，entity.Member 其余字段为零值
	FindFansMembers(memberID uint) ([]entity.Member, error)
	FindNoteLists(memberID uint) ([]entity.Note, error)
}
