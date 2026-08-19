package note

import (
	"context"
	"fmt"
	"time"

	"tuyou/internal/model/entity"
	"tuyou/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	statusCacheTTL = time.Hour
	countCacheTTL  = time.Hour
)

// Redis key 定义
func likeStatusKey(noteID, memberID uint) string {
	return fmt.Sprintf("note:like:status:%d:%d", noteID, memberID)
}
func likeCountKey(noteID uint) string { return fmt.Sprintf("note:likes:count:%d", noteID) }
func collectStatusKey(noteID, memberID uint) string {
	return fmt.Sprintf("note:collect:status:%d:%d", noteID, memberID)
}
func collectCountKey(noteID uint) string { return fmt.Sprintf("note:collects:count:%d", noteID) }

func cacheStatusValue(ok bool) string {
	if ok {
		return "1"
	}
	return "0"
}

func parseCacheStatus(value string) (bool, bool) {
	switch value {
	case "1":
		return true, true
	case "0":
		return false, true
	default:
		return false, false
	}
}

type noteRepository struct {
	db        *gorm.DB
	redisRepo repository.RedisRepository
}

// NewNoteRepository 创建笔记仓储实例
func NewNoteRepository(db *gorm.DB, redisRepo repository.RedisRepository) NoteRepository {
	return &noteRepository{db: db, redisRepo: redisRepo}
}

// Create 创建笔记
func (r *noteRepository) Create(tx *gorm.DB, note *entity.Note) error {
	return tx.Create(note).Error
}

// FindTags 根据标签名查找
func (r *noteRepository) FindTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := tx.Where("tag_name IN ? AND status = 1", tagNames).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// CreateNoteTags 批量创建笔记-标签关联
func (r *noteRepository) CreateNoteTags(tx *gorm.DB, noteID uint, tagIDs []uint) error {
	noteTags := make([]entity.NoteTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		noteTags = append(noteTags, entity.NoteTag{
			NoteID: noteID,
			TagID:  tagID,
		})
	}
	return tx.Create(&noteTags).Error
}

// CreateNoteImages 批量创建笔记图片
func (r *noteRepository) CreateNoteImages(tx *gorm.DB, noteID uint, imageURLs []string) error {
	images := make([]entity.NoteImage, 0, len(imageURLs))
	for i, url := range imageURLs {
		images = append(images, entity.NoteImage{
			NoteID: noteID,
			URL:    url,
			Sort:   i,
		})
	}
	return tx.Create(&images).Error
}

// GetAuthor 获取作者信息
func (r *noteRepository) GetAuthor(memberID uint) (*entity.Member, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// IncrPublishCount 发布计数 +1
func (r *noteRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}

// FindImagesByNoteID 查询笔记图片列表，按 sort 排序
func (r *noteRepository) FindImagesByNoteID(noteID uint) ([]entity.NoteImage, error) {
	var images []entity.NoteImage
	if err := r.db.Where("note_id = ?", noteID).
		Order("sort ASC").
		Find(&images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

// FindByID 获取笔记主表信息（仅查询上架的笔记）
func (r *noteRepository) FindByID(noteID uint) (*entity.Note, error) {
	var note entity.Note
	if err := r.db.Where("id = ? AND status = 1", noteID).First(&note).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

// FindTagsByNoteID 查询笔记的标签列表（JOIN note_tag + tag，仅启用且未删除的标签）
func (r *noteRepository) FindTagsByNoteID(noteID uint) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := r.db.Model(&entity.Tag{}).
		Joins("JOIN note_tag ON note_tag.tag_id = tag.id").
		Where("note_tag.note_id = ? AND tag.status = 1 AND tag.delete_time IS NULL", noteID).
		Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// ExistsLike 检查当前用户是否已点赞该笔记
func (r *noteRepository) ExistsLike(noteID uint, memberID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.NoteLike{}).
		Where("note_id = ? AND member_id = ?", noteID, memberID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsCollect 检查当前用户是否已收藏该笔记
func (r *noteRepository) ExistsCollect(noteID uint, memberID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.NoteCollect{}).
		Where("note_id = ? AND member_id = ?", noteID, memberID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateCollect 创建收藏记录；仅实际插入成功才同步 collect_count +1（并发重复收藏不重复计数）
func (r *noteRepository) CreateCollect(tx *gorm.DB, noteID uint, memberID uint) error {
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.NoteCollect{NoteID: noteID, MemberID: memberID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 已收藏过，不重复计数
	}
	// 作者获收藏数 +1（同事务，作者=笔记所属者）
	if err := tx.Model(&entity.Member{}).
		Where("id = (SELECT member_id FROM note WHERE id = ?)", noteID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
		return err
	}
	return tx.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("collect_count", gorm.Expr("collect_count + 1")).Error
}

// DeleteCollect 硬删除收藏记录；仅实际删除成功才同步 collect_count -1（避免并发重复取消把计数减超）
func (r *noteRepository) DeleteCollect(tx *gorm.DB, noteID uint, memberID uint) error {
	res := tx.Unscoped().Where("note_id = ? AND member_id = ?", noteID, memberID).Delete(&entity.NoteCollect{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 本来就没收藏，不减计数
	}
	// 作者获收藏数 -1（GREATEST 防减成负数）
	if err := tx.Model(&entity.Member{}).
		Where("id = (SELECT member_id FROM note WHERE id = ?)", noteID).
		UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error; err != nil {
		return err
	}
	return tx.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("collect_count", gorm.Expr("collect_count - 1")).Error
}

// UpdateStatus 更新笔记上下架状态；返回是否真正发生了状态变更
// （WHERE status <> ? 保证并发重复删除时第二个请求命中 0 行 → changed=false → 不重复扣减发布数）
func (r *noteRepository) UpdateStatus(tx *gorm.DB, noteID uint, status int) (bool, error) {
	res := tx.Model(&entity.Note{}).Where("id = ? AND status <> ?", noteID, status).
		UpdateColumn("status", status)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// SoftDelete 软删除笔记：设置 delete_time 为当前时间
func (r *noteRepository) SoftDelete(tx *gorm.DB, noteID uint) error {
	return tx.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("delete_time", time.Now()).Error
}

// DecrPublishCount 发布计数 -1
func (r *noteRepository) DecrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count - 1")).Error
}

// UpdateCommentStatusByTarget 批量更新评论状态
func (r *noteRepository) UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error {
	return tx.Model(&entity.Comment{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		UpdateColumn("status", status).Error
}

// ExistsFollow 检查当前用户是否已关注该作者
func (r *noteRepository) ExistsFollow(followerID uint, followeeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ? AND status = 1", followerID, followeeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateLike 创建点赞记录；仅实际插入成功才同步 like_count +1（并发重复点赞不重复计数）
func (r *noteRepository) CreateLike(tx *gorm.DB, noteID uint, memberID uint) error {
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&entity.NoteLike{NoteID: noteID, MemberID: memberID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 已点赞过，不重复计数
	}
	return tx.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DeleteLike 删除点赞记录（硬删除，不需要软删除）；仅实际删除成功才同步 like_count -1（避免并发重复取消把计数减超）
func (r *noteRepository) DeleteLike(tx *gorm.DB, noteID uint, memberID uint) error {
	res := tx.Unscoped().Where("note_id = ? AND member_id = ?", noteID, memberID).Delete(&entity.NoteLike{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // 本来就没点赞，不减计数
	}
	return tx.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

// IncrViewCount 浏览量 +1
func (r *noteRepository) IncrViewCount(noteID uint) error {
	return r.db.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ────────── Redis 缓存实现 ──────────

func (r *noteRepository) GetCachedLikeStatus(ctx context.Context, noteID, memberID uint) (bool, bool, error) {
	var value string
	if err := r.redisRepo.Get(ctx, likeStatusKey(noteID, memberID), &value); err != nil {
		return false, false, err
	}
	isLiked, ok := parseCacheStatus(value)
	return isLiked, ok, nil
}

func (r *noteRepository) SetCachedLikeStatus(ctx context.Context, noteID, memberID uint, isLiked bool) {
	_ = r.redisRepo.Set(ctx, likeStatusKey(noteID, memberID), cacheStatusValue(isLiked), statusCacheTTL)
}

func (r *noteRepository) GetCachedCollectStatus(ctx context.Context, noteID, memberID uint) (bool, bool, error) {
	var value string
	if err := r.redisRepo.Get(ctx, collectStatusKey(noteID, memberID), &value); err != nil {
		return false, false, err
	}
	isFavorited, ok := parseCacheStatus(value)
	return isFavorited, ok, nil
}

func (r *noteRepository) SetCachedCollectStatus(ctx context.Context, noteID, memberID uint, isFavorited bool) {
	_ = r.redisRepo.Set(ctx, collectStatusKey(noteID, memberID), cacheStatusValue(isFavorited), statusCacheTTL)
}

func (r *noteRepository) GetCachedLikeCount(ctx context.Context, noteID uint) (int64, error) {
	var count int64
	return count, r.redisRepo.Get(ctx, likeCountKey(noteID), &count)
}

func (r *noteRepository) SetCachedLikeCount(ctx context.Context, noteID uint, count int64) {
	_ = r.redisRepo.Set(ctx, likeCountKey(noteID), count, countCacheTTL)
}

func (r *noteRepository) GetCachedCollectCount(ctx context.Context, noteID uint) (int64, error) {
	var count int64
	return count, r.redisRepo.Get(ctx, collectCountKey(noteID), &count)
}

func (r *noteRepository) SetCachedCollectCount(ctx context.Context, noteID uint, count int64) {
	_ = r.redisRepo.Set(ctx, collectCountKey(noteID), count, countCacheTTL)
}

func (r *noteRepository) XAddViewCount(ctx context.Context, noteID uint) error {
	return r.redisRepo.XAdd(ctx, "stream:content", map[string]interface{}{
		"type": "note",
		"id":   noteID,
	})
}
