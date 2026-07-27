package note

import (
	responseNote "tuyou/internal/model/dto/response/web/note"
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

// GetAuthor 获取作者简要信息
func (r *noteRepository) GetAuthor(memberID uint) (*responseNote.AuthorInfo, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &responseNote.AuthorInfo{
		ID:       member.ID,
		Nickname: member.Nickname,
		Avatar:   member.Avatar,
	}, nil
}

// IncrPublishCount 发布计数 +1
func (r *noteRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}
