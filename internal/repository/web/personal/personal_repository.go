package personal

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

type personalRepository struct {
	db *gorm.DB
}

// NewPersonalRepository 创建个人中心仓储实例
func NewPersonalRepository(db *gorm.DB) PersonalRepository {
	return &personalRepository{db: db}
}

// FindFollowedMembers 查询当前用户关注的所有用户信息
//
// 通过 follow 表关联 member 表一次查出，等价于两步：
//  1. 从 follow 表按 follower_id 查到所有被关注者 ID（followee_id）
//  2. 用这些 ID 去 member 表取 nickname / avatar / bio / member_type 四列（Scan 进 entity.Member）
//
// 关联条件：
//   - follow.follower_id = 当前登录用户 ID（关注者）
//   - follow.followee_id = member.id（被关注者）
//   - follow.delete_time IS NULL 过滤已取消的关注（软删除）
//
// member 表主模型的软删除（delete_time）由 gorm 自动加上。
func (r *personalRepository) FindFollowedMembers(followerID uint) ([]entity.Member, error) {
	var rows []entity.Member
	if err := r.db.Table("member").
		Select("member.id, member.nickname, member.avatar, member.bio, member.member_type").
		Joins("JOIN follow ON follow.followee_id = member.id AND follow.delete_time IS NULL").
		Where("follow.follower_id = ?", followerID).
		Order("follow.create_time DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// FindFansMembers 查询当前用户的粉丝列表
//
// 粉丝 = 关注了我的人。此时当前用户是被关注者（followee）：
//   - follow.followee_id = 当前登录用户 ID（被关注者）
//   - follow.follower_id = member.id（关注我的人，即粉丝）
//
// 和 FindFollowedMembers 的区别：Join/Where 里 follower_id 和 followee_id 互换了。
func (r *personalRepository) FindFansMembers(memberID uint) ([]entity.Member, error) {
	var rows []entity.Member
	if err := r.db.Table("member").
		Select("member.id, member.nickname, member.avatar, member.bio, member.member_type").
		Joins("JOIN follow ON follow.follower_id = member.id AND follow.delete_time IS NULL").
		Where("follow.followee_id = ?", memberID).
		Order("follow.create_time DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *personalRepository) FindNoteLists(memberID uint) ([]entity.Note, error) {
	var rows []entity.Note

}
