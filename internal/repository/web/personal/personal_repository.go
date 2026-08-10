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

// FindFollowedMembers 分页查询当前用户关注的所有用户信息
func (r *personalRepository) FindFollowedMembers(followerID uint, page, size int) ([]entity.Member, int64, error) {
	var total int64
	if err := r.db.Table("member").
		Joins("JOIN follow ON follow.followee_id = member.id AND follow.delete_time IS NULL").
		Where("follow.follower_id = ?", followerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []entity.Member
	if err := r.db.Table("member").
		Select("member.id, member.nickname, member.avatar, member.bio, member.role").
		Joins("JOIN follow ON follow.followee_id = member.id AND follow.delete_time IS NULL").
		Where("follow.follower_id = ?", followerID).
		Order("follow.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// FindFansMembers 分页查询当前用户的粉丝列表
func (r *personalRepository) FindFansMembers(memberID uint, page, size int) ([]entity.Member, int64, error) {
	var total int64
	if err := r.db.Table("member").
		Joins("JOIN follow ON follow.follower_id = member.id AND follow.delete_time IS NULL").
		Where("follow.followee_id = ?", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []entity.Member
	if err := r.db.Table("member").
		Select("member.id, member.nickname, member.avatar, member.bio, member.role").
		Joins("JOIN follow ON follow.follower_id = member.id AND follow.delete_time IS NULL").
		Where("follow.followee_id = ?", memberID).
		Order("follow.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// FindFollowedIDs 查询当前用户关注的所有用户 ID 集合
func (r *personalRepository) FindFollowedIDs(followerID uint) (map[uint]bool, error) {
	var ids []uint
	if err := r.db.Table("follow").
		Select("followee_id").
		Where("follower_id = ? AND delete_time IS NULL", followerID).
		Pluck("followee_id", &ids).Error; err != nil {
		return nil, err
	}

	followed := make(map[uint]bool, len(ids))
	for _, id := range ids {
		followed[id] = true
	}
	return followed, nil
}

func (r *personalRepository) CreateFollow(followerID, followeeID uint) error {
	var count int64
	if err := r.db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return r.db.Create(&entity.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}).Error
}

// DeleteFollow 取关：硬删除，返回是否真的删掉了（true=之前处于已关注状态）
func (r *personalRepository) DeleteFollow(followerID, followeeID uint) (bool, error) {
	res := r.db.Unscoped().
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&entity.Follow{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *personalRepository) FindNoteLists(memberID uint, page, size int) ([]entity.Note, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Note{}).
		Where("member_id = ? AND status = 1", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var notes []entity.Note
	if err := r.db.Where("member_id = ? AND status = 1", memberID).
		Order("publish_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&notes).Error; err != nil {
		return nil, total, err
	}
	return notes, total, nil
}
