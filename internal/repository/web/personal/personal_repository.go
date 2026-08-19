package personal

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type personalRepository struct {
	db *gorm.DB
}

// NewPersonalRepository 创建个人中心仓储实例
func NewPersonalRepository(db *gorm.DB) PersonalRepository {
	return &personalRepository{db: db}
}

// 个人信息
func (r *personalRepository) FindByID(id uint) (*entity.Member, error) {
	var member entity.Member
	if err := r.db.First(&member, id).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *personalRepository) UpdateProfile(memberID uint, updates map[string]interface{}) error {
	return r.db.Model(&entity.Member{}).Where("id = ?", memberID).Updates(updates).Error
}

// FindFollowedMembers 分页查询当前用户关注的所有用户信息
func (r *personalRepository) FindFollowedMembers(followerID uint, page, size int) ([]entity.Member, int64, error) {
	var total int64
	if err := r.db.Table("member").
		Joins("JOIN follow ON follow.followee_id = member.id AND follow.delete_time IS NULL AND follow.status = 1").
		Where("follow.follower_id = ? AND member.delete_time IS NULL", followerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []entity.Member
	if err := r.db.Table("member").
		Select("member.id, member.nickname, member.avatar, member.bio, member.role").
		Joins("JOIN follow ON follow.followee_id = member.id AND follow.delete_time IS NULL AND follow.status = 1").
		Where("follow.follower_id = ? AND member.delete_time IS NULL", followerID).
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
		Joins("JOIN follow ON follow.follower_id = member.id AND follow.delete_time IS NULL AND follow.status = 1").
		Where("follow.followee_id = ? AND member.delete_time IS NULL", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []entity.Member
	if err := r.db.Table("member").
		Select("member.id, member.nickname, member.avatar, member.bio, member.role").
		Joins("JOIN follow ON follow.follower_id = member.id AND follow.delete_time IS NULL AND follow.status = 1").
		Where("follow.followee_id = ? AND member.delete_time IS NULL", memberID).
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
		Where("follower_id = ? AND delete_time IS NULL AND status = 1", followerID).
		Pluck("followee_id", &ids).Error; err != nil {
		return nil, err
	}

	followed := make(map[uint]bool, len(ids))
	for _, id := range ids {
		followed[id] = true
	}
	return followed, nil
}

// ToggleFollow 原子翻转关注状态：行不存在 → 插入 status=1（关注）；行已存在 → status = 1-status（翻转）。
// 唯一索引 idx_follower_followee 作并发串行器：两个并发请求由 InnoDB 强制排队，
// 后到者走 ON DUPLICATE KEY UPDATE 翻转，杜绝"删0行误走关注分支"和撞索引 500。
// 返回 (翻转后是否已关注, 计数增量 delta)：已关注 → +1，已取关 → -1。
func (r *personalRepository) ToggleFollow(tx *gorm.DB, followerID, followeeID uint) (bool, int, error) {
	res := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "follower_id"}, {Name: "followee_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status": gorm.Expr("1 - status"),
		}),
	}).Create(&entity.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		Status:     1,
	})
	if res.Error != nil {
		return false, 0, res.Error
	}

	// 翻转结果需回读（MySQL 无 RETURNING）：同一事务内行必已存在
	var f entity.Follow
	if err := tx.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&f).Error; err != nil {
		return false, 0, err
	}
	if f.Status == 1 {
		return true, 1, nil
	}
	return false, -1, nil
}

// AdjustFollowCount 我的关注数 +delta（GREATEST 防负，P2）
func (r *personalRepository) AdjustFollowCount(tx *gorm.DB, memberID uint, delta int) error {
	return tx.Model(&entity.Member{}).
		Where("id = ?", memberID).
		UpdateColumn("follow_count", gorm.Expr("GREATEST(follow_count + ?, 0)", delta)).Error
}

// AdjustFansCount 对方粉丝数 +delta（GREATEST 防负，P2）
func (r *personalRepository) AdjustFansCount(tx *gorm.DB, memberID uint, delta int) error {
	return tx.Model(&entity.Member{}).
		Where("id = ?", memberID).
		UpdateColumn("fans_count", gorm.Expr("GREATEST(fans_count + ?, 0)", delta)).Error
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

// FindLikedStrategys 分页查询当前用户点赞过的攻略，按点赞时间倒序（仅上架）
func (r *personalRepository) FindLikedStrategys(memberID uint, page, size int) ([]entity.Strategy, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Strategy{}).
		Joins("JOIN strategy_like ON strategy_like.strategy_id = strategy.id").
		Where("strategy_like.member_id = ? AND strategy.status = 1", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var strategys []entity.Strategy
	if err := r.db.Model(&entity.Strategy{}).
		Joins("JOIN strategy_like ON strategy_like.strategy_id = strategy.id").
		Where("strategy_like.member_id = ? AND strategy.status = 1", memberID).
		Order("strategy_like.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&strategys).Error; err != nil {
		return nil, total, err
	}
	return strategys, total, nil
}

// FindCollectStrategys 分页查询当前用户收藏过的攻略，按收藏时间倒序（仅上架）
func (r *personalRepository) FindCollectStrategys(memberID uint, page, size int) ([]entity.Strategy, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Strategy{}).
		Joins("JOIN strategy_collect ON strategy_collect.strategy_id = strategy.id").
		Where("strategy_collect.member_id = ? AND strategy.status = 1", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var strategys []entity.Strategy
	if err := r.db.Model(&entity.Strategy{}).
		Joins("JOIN strategy_collect ON strategy_collect.strategy_id = strategy.id").
		Where("strategy_collect.member_id = ? AND strategy.status = 1", memberID).
		Order("strategy_collect.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&strategys).Error; err != nil {
		return nil, total, err
	}
	return strategys, total, nil
}

// FindPublishStrategys 分页查询当前用户发布的攻略，按发布时间倒序（仅上架）
func (r *personalRepository) FindPublishStrategys(memberID uint, page, size int) ([]entity.Strategy, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Strategy{}).
		Where("member_id = ? AND status = 1", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var strategys []entity.Strategy
	if err := r.db.Where("member_id = ? AND status = 1", memberID).
		Order("publish_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&strategys).Error; err != nil {
		return nil, total, err
	}
	return strategys, total, nil
}

// FindCollectNotes 分页查询当前用户收藏过的笔记，按收藏时间倒序（仅上架）
func (r *personalRepository) FindCollectNotes(memberID uint, page, size int) ([]entity.Note, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Note{}).
		Joins("JOIN note_collect ON note_collect.note_id = note.id").
		Where("note_collect.member_id = ? AND note.status = 1", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notes []entity.Note
	if err := r.db.Model(&entity.Note{}).
		Joins("JOIN note_collect ON note_collect.note_id = note.id").
		Where("note_collect.member_id = ? AND note.status = 1", memberID).
		Order("note_collect.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&notes).Error; err != nil {
		return nil, total, err
	}
	return notes, total, nil
}

// FindLikedNotes 分页查询当前用户点赞过的笔记，按点赞时间倒序（仅上架）
func (r *personalRepository) FindLikedNotes(memberID uint, page, size int) ([]entity.Note, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Note{}).
		Joins("JOIN note_like ON note_like.note_id = note.id").
		Where("note_like.member_id = ? AND note.status = 1", memberID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notes []entity.Note
	if err := r.db.Model(&entity.Note{}).
		Joins("JOIN note_like ON note_like.note_id = note.id").
		Where("note_like.member_id = ? AND note.status = 1", memberID).
		Order("note_like.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&notes).Error; err != nil {
		return nil, total, err
	}
	return notes, total, nil
}

// FindPublishActivities 分页查询当前用户发布的活动（未删除，按开始时间倒序）
func (r *personalRepository) FindPublishActivities(memberID uint, page, size int) ([]entity.Activity, int64, error) {
	var total int64
	db := r.db.Model(&entity.Activity{}).
		Where("leader_member_id = ? AND is_del = ?", memberID, 0)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []entity.Activity
	if err := db.Offset((page - 1) * size).Limit(size).
		Order("start_time DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// FindJoinActivities 分页查询当前用户报名参加的活动（未取消、活动未删除，按报名时间倒序）
func (r *personalRepository) FindJoinActivities(memberID uint, page, size int) ([]entity.Activity, int64, error) {
	var total int64
	if err := r.db.Model(&entity.Activity{}).
		Joins("JOIN activity_join ON activity_join.activity_id = activity.id").
		Where("activity_join.member_id = ? AND activity_join.is_cancel = ? AND activity.is_del = ?", memberID, 0, 0).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []entity.Activity
	if err := r.db.Model(&entity.Activity{}).
		Joins("JOIN activity_join ON activity_join.activity_id = activity.id").
		Where("activity_join.member_id = ? AND activity_join.is_cancel = ? AND activity.is_del = ?", memberID, 0, 0).
		Order("activity_join.create_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ========== 卡片批量查询（避免 N+1，仅本模块使用，不改动 note/guide 仓储） ==========

// FindNoteImagesByIDs 批量查询多个笔记的图片，按 sort 升序（返回 noteID -> 图片列表）
func (r *personalRepository) FindNoteImagesByIDs(noteIDs []uint) (map[uint][]entity.NoteImage, error) {
	result := make(map[uint][]entity.NoteImage, len(noteIDs))
	if len(noteIDs) == 0 {
		return result, nil
	}
	var images []entity.NoteImage
	if err := r.db.Where("note_id IN ?", noteIDs).
		Order("sort ASC").
		Find(&images).Error; err != nil {
		return nil, err
	}
	for _, img := range images {
		result[img.NoteID] = append(result[img.NoteID], img)
	}
	return result, nil
}

// FindNoteTagsByIDs 批量查询多个笔记的标签（返回 noteID -> 标签列表）
func (r *personalRepository) FindNoteTagsByIDs(noteIDs []uint) (map[uint][]entity.Tag, error) {
	result := make(map[uint][]entity.Tag, len(noteIDs))
	if len(noteIDs) == 0 {
		return result, nil
	}

	// 1. 先查 note_tag 关联表，得到 noteID -> tagIDs
	var noteTags []entity.NoteTag
	if err := r.db.Where("note_id IN ?", noteIDs).Find(&noteTags).Error; err != nil {
		return nil, err
	}
	tagIDSet := make(map[uint]bool)
	tagIDs := make([]uint, 0, len(noteTags))
	noteTagMap := make(map[uint][]uint, len(noteTags))
	for _, nt := range noteTags {
		noteTagMap[nt.NoteID] = append(noteTagMap[nt.NoteID], nt.TagID)
		if !tagIDSet[nt.TagID] {
			tagIDSet[nt.TagID] = true
			tagIDs = append(tagIDs, nt.TagID)
		}
	}

	// 2. 再一次性查标签
	var tags []entity.Tag
	if err := r.db.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
		return nil, err
	}
	tagMap := make(map[uint]entity.Tag, len(tags))
	for _, t := range tags {
		tagMap[t.ID] = t
	}

	for noteID, ids := range noteTagMap {
		for _, tagID := range ids {
			if t, ok := tagMap[tagID]; ok {
				result[noteID] = append(result[noteID], t)
			}
		}
	}
	return result, nil
}

// FindStrategyImagesByIDs 批量查询多个攻略的图片，按 sort 升序（返回 strategyID -> 图片列表）
func (r *personalRepository) FindStrategyImagesByIDs(strategyIDs []uint) (map[uint][]entity.StrategyImage, error) {
	result := make(map[uint][]entity.StrategyImage, len(strategyIDs))
	if len(strategyIDs) == 0 {
		return result, nil
	}
	var images []entity.StrategyImage
	if err := r.db.Where("strategy_id IN ?", strategyIDs).
		Order("sort ASC").
		Find(&images).Error; err != nil {
		return nil, err
	}
	for _, img := range images {
		result[img.StrategyID] = append(result[img.StrategyID], img)
	}
	return result, nil
}

// FindStrategyTagsByIDs 批量查询多个攻略的标签（返回 strategyID -> 标签列表）
func (r *personalRepository) FindStrategyTagsByIDs(strategyIDs []uint) (map[uint][]entity.Tag, error) {
	result := make(map[uint][]entity.Tag, len(strategyIDs))
	if len(strategyIDs) == 0 {
		return result, nil
	}

	// 1. 先查 strategy_tag 关联表，得到 strategyID -> tagIDs
	var strategyTags []entity.StrategyTag
	if err := r.db.Where("strategy_id IN ?", strategyIDs).Find(&strategyTags).Error; err != nil {
		return nil, err
	}
	tagIDSet := make(map[uint]bool)
	tagIDs := make([]uint, 0, len(strategyTags))
	strategyTagMap := make(map[uint][]uint, len(strategyTags))
	for _, st := range strategyTags {
		strategyTagMap[st.StrategyID] = append(strategyTagMap[st.StrategyID], st.TagID)
		if !tagIDSet[st.TagID] {
			tagIDSet[st.TagID] = true
			tagIDs = append(tagIDs, st.TagID)
		}
	}

	// 2. 再一次性查标签
	var tags []entity.Tag
	if err := r.db.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
		return nil, err
	}
	tagMap := make(map[uint]entity.Tag, len(tags))
	for _, t := range tags {
		tagMap[t.ID] = t
	}

	for strategyID, ids := range strategyTagMap {
		for _, tagID := range ids {
			if t, ok := tagMap[tagID]; ok {
				result[strategyID] = append(result[strategyID], t)
			}
		}
	}
	return result, nil
}

// FindMembersByIDs 批量查询会员信息（走 Model 自动过滤已删除用户）
func (r *personalRepository) FindMembersByIDs(memberIDs []uint) ([]entity.Member, error) {
	if len(memberIDs) == 0 {
		return nil, nil
	}
	var members []entity.Member
	if err := r.db.Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// FindLikedNoteIDs 批量查询用户点赞过的笔记ID集合
func (r *personalRepository) FindLikedNoteIDs(noteIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(noteIDs))
	if len(noteIDs) == 0 {
		return result, nil
	}
	var ids []uint
	if err := r.db.Model(&entity.NoteLike{}).
		Where("note_id IN ? AND member_id = ?", noteIDs, memberID).
		Pluck("note_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

// FindLikedStrategyIDs 批量查询用户点赞过的攻略ID集合
func (r *personalRepository) FindLikedStrategyIDs(strategyIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(strategyIDs))
	if len(strategyIDs) == 0 {
		return result, nil
	}
	var ids []uint
	if err := r.db.Model(&entity.StrategyLike{}).
		Where("strategy_id IN ? AND member_id = ?", strategyIDs, memberID).
		Pluck("strategy_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		result[id] = true
	}
	return result, nil
}

// FindParticipantsByActivityIDs 批量查询多个活动的报名人（activity_join JOIN member）
// 返回 activityID -> 报名用户列表，按报名时间升序；手机号字段由调用方保证只用于"领队本人"可见出口。
func (r *personalRepository) FindParticipantsByActivityIDs(activityIDs []uint) (map[uint][]entity.Member, error) {
	result := make(map[uint][]entity.Member)
	if len(activityIDs) == 0 {
		return result, nil
	}
	type joinRow struct {
		ActivityID uint
		entity.Member
	}
	var rows []joinRow
	if err := r.db.Table("activity_join").
		Select("activity_join.activity_id, member.id, member.nickname, member.avatar, member.phone").
		Joins("JOIN member ON member.id = activity_join.member_id").
		Where("activity_join.activity_id IN ? AND activity_join.is_cancel = 0 AND member.delete_time IS NULL", activityIDs).
		Order("activity_join.create_time ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ActivityID] = append(result[row.ActivityID], row.Member)
	}
	return result, nil
}
