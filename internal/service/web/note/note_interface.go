package note

import (
	requestNote "tuyou/internal/model/dto/request/web/note"
	responseNote "tuyou/internal/model/dto/response/web/note"
	"tuyou/internal/model/entity"
)

// NoteService 笔记服务接口
type NoteService interface {
	// CreateNote 发布笔记
	CreateNote(memberID uint, req *requestNote.CreateNoteRequest) (*responseNote.CreateNoteResponse, error)

	// GetNoteDetail 获取笔记详情
	GetNoteDetail(noteID uint, memberID uint) (*responseNote.NoteDetailResponse, error)

	// ToggleLikeNote 点赞/取消点赞
	ToggleLikeNote(noteID uint, memberID uint) (*responseNote.LikeToggleResponse, error)

	// ToggleCollectNote 收藏/取消收藏
	ToggleCollectNote(noteID uint, memberID uint) (*responseNote.CollectToggleResponse, error)

	// DeleteNote 删除笔记（下架，仅作者）
	DeleteNote(noteID uint, memberID uint) error

	BuildNoteCards(notes []entity.Note, memberID uint) ([]responseNote.NoteCard, error)

	// GetCollectNotes 当前用户收藏的笔记列表（分页）
	GetCollectNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error)

	// GetLikedNotes 当前用户点赞的笔记列表（分页）
	GetLikedNotes(memberID uint, page, size int) ([]responseNote.NoteCard, int64, error)
}
