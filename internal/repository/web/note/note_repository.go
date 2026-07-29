package note

import (
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

type noteRepository struct {
	db *gorm.DB
}

// NewNoteRepository 创建笔记仓储实例
func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
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

// FindTagsByNoteID 查询笔记的标签列表（JOIN note_tag + tag）
func (r *noteRepository) FindTagsByNoteID(noteID uint) ([]entity.Tag, error) {
	var tags []entity.Tag
	if err := r.db.Model(&entity.Tag{}).
		Joins("JOIN note_tag ON note_tag.tag_id = tag.id").
		Where("note_tag.note_id = ?", noteID).
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

// ExistsFollow 检查当前用户是否已关注该作者
func (r *noteRepository) ExistsFollow(followerID uint, followeeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.Follow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// IncrViewCount 浏览量 +1
func (r *noteRepository) IncrViewCount(noteID uint) error {
	return r.db.Model(&entity.Note{}).Where("id = ?", noteID).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
