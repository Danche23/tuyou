package note

import (
	responseNote "tuyou/internal/model/dto/response/web/note"
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// NoteRepository 笔记仓储接口
type NoteRepository interface {
	// Create 创建笔记
	Create(tx *gorm.DB, note *entity.Note) error

	// FindTags 根据标签名查找
	FindTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error)

	// CreateNoteTags 批量创建笔记-标签关联
	CreateNoteTags(tx *gorm.DB, noteID uint, tagIDs []uint) error

	// CreateNoteImages 批量创建笔记图片
	CreateNoteImages(tx *gorm.DB, noteID uint, imageURLs []string) error

	// GetAuthor 获取作者简要信息
	GetAuthor(memberID uint) (*responseNote.AuthorInfo, error)

	// IncrPublishCount 发布计数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error
}
