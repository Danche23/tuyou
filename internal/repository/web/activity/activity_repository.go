package activity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tuyou/internal/model/entity"
	"tuyou/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type activityRepository struct {
	db        *gorm.DB
	redisRepo repository.RedisRepository
}

// NewActivityRepository 创建活动仓储实例
func NewActivityRepository(db *gorm.DB, redisRepo repository.RedisRepository) ActivityRepository {
	return &activityRepository{db: db, redisRepo: redisRepo}
}

// Create 创建活动
func (r *activityRepository) Create(tx *gorm.DB, activity *entity.Activity) error {
	return tx.Create(activity).Error
}

// GetAuthor 获取作者（领队）信息
func (r *activityRepository) GetAuthor(memberID uint) (*entity.Member, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// FindByID 根据活动ID查询（未删除）
func (r *activityRepository) FindByID(activityID uint) (*entity.Activity, error) {
	var activity entity.Activity
	if err := r.db.Where("id = ? AND is_del = ?", activityID, 0).First(&activity).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

// FindJoinedMembers 查询已报名成员列表（activity_join JOIN member，仅有效报名）
func (r *activityRepository) FindJoinedMembers(activityID uint) ([]entity.Member, error) {
	var members []entity.Member
	if err := r.db.Table("activity_join").
		Select("member.id, member.nickname, member.avatar").
		Joins("JOIN member ON member.id = activity_join.member_id").
		Where("activity_join.activity_id = ? AND activity_join.is_cancel = 0 AND member.delete_time IS NULL", activityID).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// ExistsJoin 检查用户是否已报名该活动（仅计算有效报名，is_cancel=0）
func (r *activityRepository) ExistsJoin(activityID uint, memberID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.ActivityJoin{}).
		Where("activity_id = ? AND member_id = ? AND is_cancel = 0", activityID, memberID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsFollow 检查当前用户是否已关注领队
func (r *activityRepository) ExistsFollow(followerID uint, followeeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ? AND status = 1", followerID, followeeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateActivityStatus 更新活动状态
func (r *activityRepository) UpdateActivityStatus(activityID uint, status int) error {
	return r.db.Model(&entity.Activity{}).Where("id = ?", activityID).
		UpdateColumn("activity_status", status).Error
}

// CreateJoin 创建报名记录。
// 若已有取消记录（is_cancel=1）则恢复报名（is_cancel=0），否则新建记录。
// 并发双击时唯一索引兜底：第二次插入被忽略，返回 false，事务外层不再加人数、不发通知。
func (r *activityRepository) CreateJoin(tx *gorm.DB, activityID uint, memberID uint) (bool, error) {
	// 先尝试恢复已取消的记录
	res := tx.Model(&entity.ActivityJoin{}).
		Where("activity_id = ? AND member_id = ? AND is_cancel = 1", activityID, memberID).
		UpdateColumn("is_cancel", 0)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected > 0 {
		return true, nil // 恢复成功
	}

	// 没有取消记录，尝试新建
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entity.ActivityJoin{ActivityID: activityID, MemberID: memberID})
	return result.RowsAffected > 0, result.Error
}

// DeleteJoin 取消报名（软删：is_cancel=1，cancel_count+1）；返回是否实际取消了记录
func (r *activityRepository) DeleteJoin(tx *gorm.DB, activityID uint, memberID uint) (bool, error) {
	res := tx.Model(&entity.ActivityJoin{}).
		Where("activity_id = ? AND member_id = ? AND is_cancel = 0", activityID, memberID).
		Updates(map[string]interface{}{
			"is_cancel":    1,
			"cancel_count": gorm.Expr("cancel_count + 1"),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// IncrCurrentPeople 报名人数 +1，返回实际影响行数（0=满员）
func (r *activityRepository) IncrCurrentPeople(tx *gorm.DB, activityID uint, maxPeople int) (int64, error) {
	result := tx.Model(&entity.Activity{}).
		Where("id = ? AND current_people < ?", activityID, maxPeople).
		UpdateColumn("current_people", gorm.Expr("current_people + 1"))
	return result.RowsAffected, result.Error
}

// DecrCurrentPeople 报名人数 -1（加 current_people>0 下限保护，防并发/异常下被减成负数）
func (r *activityRepository) DecrCurrentPeople(tx *gorm.DB, activityID uint) error {
	return tx.Model(&entity.Activity{}).Where("id = ? AND current_people > 0", activityID).
		UpdateColumn("current_people", gorm.Expr("current_people - 1")).Error
}

// signupSlotsKey 报名剩余名额闸门 key（Redis 预扣减；DB 的 current_people 仍是最终裁判）
func signupSlotsKey(activityID uint) string {
	return fmt.Sprintf("activity:signup:slots:%d", activityID)
}

// InitSignupSlots 创建活动时初始化剩余名额（best-effort；Redis 不可用则跳过，靠闸门缺失回填自愈）
func (r *activityRepository) InitSignupSlots(ctx context.Context, activityID uint, slots int) {
	_ = r.redisRepo.Set(ctx, signupSlotsKey(activityID), int64(slots), 0)
}

// GateSignupSlot 报名闸门（Redis 预扣减）：先 DECR 抢名额，再决定是否放行。
//   - allowed=true, acquired=true：在 Redis 扣到名额，放行；若后续 DB 事务未成功（重复报名/满员/失败）须调用 ReleaseSignupSlot 归还。
//   - allowed=true, acquired=false：Redis 不可用（降级），未在 Redis 扣名额，直接交给 DB 乐观锁兜底，无需归还。
//   - allowed=false：名额已满（含冷启动按 DB 回填后仍满），不放行。
//
// 冷启动/漂移自愈：DECR<0 时按 DB 真实剩余名额回填后再扣一次；DB 的 WHERE current_people<max_people 始终做最终防超卖裁判。
func (r *activityRepository) GateSignupSlot(ctx context.Context, activityID uint, maxPeople int) (allowed bool, acquired bool, err error) {
	key := signupSlotsKey(activityID)
	slots, err := r.redisRepo.IncrBy(ctx, key, -1)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return true, false, nil // Redis 不可用：降级，不闸门
		}
		return false, false, err
	}
	if slots >= 0 {
		return true, true, nil // 抢到名额
	}
	// slots<0：可能真满，也可能冷启动（key 缺失被扣成负数）。先归还这次多扣，再按 DB 判断
	_, _ = r.redisRepo.IncrBy(ctx, key, +1)

	current, err := r.GetCurrentPeople(r.db, activityID)
	if err != nil {
		return false, false, err
	}
	if current >= maxPeople {
		return false, false, nil // 真满
	}
	// DB 还有名额但 Redis 判满 → 冷启动/漂移：回填剩余名额后再扣一次
	_ = r.redisRepo.Set(ctx, key, int64(maxPeople-current), 0)
	slots, err = r.redisRepo.IncrBy(ctx, key, -1)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return true, false, nil
		}
		return false, false, err
	}
	if slots >= 0 {
		return true, true, nil
	}
	_, _ = r.redisRepo.IncrBy(ctx, key, +1)
	return false, false, nil
}

// ReleaseSignupSlot 归还一个名额（报名回滚/取消报名时调用）。
// key 不存在时不回填，避免冷启动把名额写错（由闸门缺失回填自愈）。
func (r *activityRepository) ReleaseSignupSlot(ctx context.Context, activityID uint) {
	key := signupSlotsKey(activityID)
	var v int64
	if err := r.redisRepo.Get(ctx, key, &v); err != nil {
		return // key 不存在或 Redis 不可用：跳过
	}
	_, _ = r.redisRepo.IncrBy(ctx, key, +1)
}

// GetCurrentPeople 查询活动当前报名人数（返回 DB 真实值，供报名/取消后响应使用）
func (r *activityRepository) GetCurrentPeople(db *gorm.DB, activityID uint) (int, error) {
	var activity entity.Activity
	if err := db.Select("current_people").First(&activity, activityID).Error; err != nil {
		return 0, err
	}
	return activity.CurrentPeople, nil
}

// SoftDelete 逻辑删除活动（is_del=1 + delete_time，与笔记/攻略软删语义对齐）
func (r *activityRepository) SoftDelete(tx *gorm.DB, activityID uint) error {
	return tx.Unscoped().Model(&entity.Activity{}).Where("id = ?", activityID).
		Updates(map[string]interface{}{
			"is_del":      1,
			"delete_time": time.Now(),
		}).Error
}

// FindList 查询活动列表（分页，按分类/日期筛选，过滤已删除，按开始时间倒序）
// dateFilter 为零值时不启用日期筛选；非零值时筛选 start_time <= 当天23:59:59 AND end_time >= 当天00:00:00 的活动
func (r *activityRepository) FindList(category int, page, size int, dateFilter time.Time) ([]entity.Activity, int64, error) {
	var list []entity.Activity
	var total int64

	db := r.db.Model(&entity.Activity{}).Where("is_del = ?", 0)
	if category > 0 {
		db = db.Where("activity_type = ?", category)
	}
	if !dateFilter.IsZero() {
		dayStart := time.Date(dateFilter.Year(), dateFilter.Month(), dateFilter.Day(), 0, 0, 0, 0, time.Local)
		dayEnd := dayStart.Add(24*time.Hour - time.Second)
		db = db.Where("start_time <= ? AND end_time >= ?", dayEnd, dayStart)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Offset(offset).Limit(size).Order("start_time DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetMembersByIDs 批量获取成员信息
func (r *activityRepository) GetMembersByIDs(ids []uint) ([]entity.Member, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var members []entity.Member
	if err := r.db.Where("id IN ?", ids).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// FindJoinedActivityIDs 查询当前用户已报名的活动ID集合（仅有效报名）
func (r *activityRepository) FindJoinedActivityIDs(activityIDs []uint, memberID uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if len(activityIDs) == 0 {
		return result, nil
	}

	var joins []entity.ActivityJoin
	if err := r.db.Model(&entity.ActivityJoin{}).
		Where("activity_id IN ? AND member_id = ? AND is_cancel = 0", activityIDs, memberID).
		Find(&joins).Error; err != nil {
		return nil, err
	}

	for _, j := range joins {
		result[j.ActivityID] = true
	}
	return result, nil
}

// GetCancelCount 查询用户对某活动的累计取消次数
func (r *activityRepository) GetCancelCount(activityID uint, memberID uint) (int, error) {
	var join entity.ActivityJoin
	if err := r.db.Select("cancel_count").
		Where("activity_id = ? AND member_id = ?", activityID, memberID).
		First(&join).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return join.CancelCount, nil
}
