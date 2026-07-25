package service

import (
	"tuyou/internal/model/dto/request"
	"tuyou/internal/model/dto/response"
)

// NoteService 笔记服务接口
type NoteService interface {
	// Create 发布笔记
	Create(memberID uint, req *request.CreateNoteRequest) (*response.NoteResponse, error)
}
