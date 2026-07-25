package repository

import (
	"tuyou/internal/model/dto/response"
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

// FindOrCreateTags 根据标签名查找或创建标签
func (r *noteRepository) FindOrCreateTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error) {
	tags := make([]entity.Tag, 0, len(tagNames))
	for _, name := range tagNames {
		var tag entity.Tag
		err := tx.Where("tag_name = ?", name).First(&tag).Error
		if err == gorm.ErrRecordNotFound {
			tag = entity.Tag{TagName: name, Status: 1}
			if err := tx.Create(&tag).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
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
func (r *noteRepository) GetAuthor(memberID uint) (*response.AuthorInfo, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &response.AuthorInfo{
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
