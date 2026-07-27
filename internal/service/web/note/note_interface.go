package note

import (
	requestNote "tuyou/internal/model/dto/request/web/note"
	responseNote "tuyou/internal/model/dto/response/web/note"
)

// NoteService 笔记服务接口
type NoteService interface {
	// Create 发布笔记
	Create(memberID uint, req *requestNote.CreateNoteRequest) (*responseNote.CreateNoteResponse, error)
}
