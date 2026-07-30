package note

import (
	requestNote "tuyou/internal/model/dto/request/web/note"
	responseNote "tuyou/internal/model/dto/response/web/note"
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
}
