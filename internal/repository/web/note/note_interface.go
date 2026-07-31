package note

import (
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

	// GetAuthor 获取作者信息
	GetAuthor(memberID uint) (*entity.Member, error)

	// IncrPublishCount 发布计数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error

	// FindByID 根据笔记ID查笔记主表
	FindByID(noteID uint) (*entity.Note, error)

	// FindImagesByNoteID 查询笔记图片列表
	FindImagesByNoteID(noteID uint) ([]entity.NoteImage, error)

	// FindTagsByNoteID 查询笔记的标签列表（JOIN note_tag + tag）
	FindTagsByNoteID(noteID uint) ([]entity.Tag, error)

	// ExistsLike 检查用户是否已点赞该笔记
	ExistsLike(noteID uint, memberID uint) (bool, error)

	// ExistsCollect 检查用户是否已收藏该笔记
	ExistsCollect(noteID uint, memberID uint) (bool, error)

	// CreateCollect 创建收藏记录
	CreateCollect(tx *gorm.DB, noteID uint, memberID uint) error

	// DeleteCollect 删除收藏记录（硬删除）
	DeleteCollect(tx *gorm.DB, noteID uint, memberID uint) error

	// IncrCollectCount 笔记收藏数 +1
	IncrCollectCount(tx *gorm.DB, noteID uint) error

	// DecrCollectCount 笔记收藏数 -1
	DecrCollectCount(tx *gorm.DB, noteID uint) error

	// UpdateStatus 更新笔记上下架状态
	UpdateStatus(tx *gorm.DB, noteID uint, status int) error

	// DecrPublishCount 发布计数 -1
	DecrPublishCount(tx *gorm.DB, memberID uint) error

	// UpdateCommentStatusByTarget 批量更新评论状态
	UpdateCommentStatusByTarget(tx *gorm.DB, targetType int, targetID uint, status int) error

	// ExistsFollow 检查用户是否已关注该作者
	ExistsFollow(followerID uint, followeeID uint) (bool, error)

	// IncrViewCount 浏览量 +1
	IncrViewCount(noteID uint) error

	// CreateLike 创建点赞记录
	CreateLike(tx *gorm.DB, noteID uint, memberID uint) error

	// DeleteLike 删除点赞记录
	DeleteLike(tx *gorm.DB, noteID uint, memberID uint) error

	// IncrLikeCount 笔记点赞数 +1
	IncrLikeCount(tx *gorm.DB, noteID uint) error

	// DecrLikeCount 笔记点赞数 -1
	DecrLikeCount(tx *gorm.DB, noteID uint) error
}
