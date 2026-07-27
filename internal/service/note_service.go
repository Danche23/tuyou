package service

import (
	"time"

	"tuyou/internal/model/dto/request"
	"tuyou/internal/model/dto/response"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	"tuyou/pkg/errors"

	"gorm.io/gorm"
)

type noteService struct {
	noteRepo repository.NoteRepository
	db       *gorm.DB
}

// NewNoteService 创建笔记服务实例
func NewNoteService(noteRepo repository.NoteRepository, db *gorm.DB) NoteService {
	return &noteService{noteRepo: noteRepo, db: db}
}

// Create 发布笔记
func (s *noteService) Create(memberID uint, req *request.CreateNoteRequest) (*response.NoteResponse, error) {
	// ========== 输入校验（事务外，快失败） ==========

	// 校验图片/视频互斥
	if req.VideoURL != "" && len(req.Images) != 1 {
		return nil, errors.New(errors.CodeInvalidParam, "视频笔记只能上传1张封面图片")
	}

	// 校验标签有效性
	tags, err := s.noteRepo.FindTags(s.db, req.Tags)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "标签查询失败", err)
	}
	if len(tags) != len(req.Tags) {
		return nil, errors.New(errors.CodeInvalidParam, "标签不存在或已禁用")
	}

	// ========== 写入（事务内） ==========

	// 构建 Note 实体
	note := &entity.Note{
		MemberID:    memberID,
		Title:       req.Title,
		Content:     req.Content,
		Category:    req.Category,
		VideoURL:    req.VideoURL,
		Status:      1,
		PublishTime: time.Now(),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建笔记
		if err := s.noteRepo.Create(tx, note); err != nil {
			return err
		}

		// 2. 创建笔记-标签关联
		tagIDs := make([]uint, len(tags))
		for i, t := range tags {
			tagIDs[i] = t.ID
		}
		if err := s.noteRepo.CreateNoteTags(tx, note.ID, tagIDs); err != nil {
			return err
		}

		// 3. 创建图片
		if err := s.noteRepo.CreateNoteImages(tx, note.ID, req.Images); err != nil {
			return err
		}

		// 4. 更新发布计数
		if err := s.noteRepo.IncrPublishCount(tx, memberID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "发布笔记失败", err)
	}

	// ========== 组装响应（事务外，避免长事务） ==========

	author, err := s.noteRepo.GetAuthor(memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}

	coverImg := ""
	if len(req.Images) > 0 {
		coverImg = req.Images[0]
	}
	return &response.NoteResponse{
		ID:           note.ID,
		Type:         "note",
		Author:       author,
		Title:        note.Title,
		CoverImg:     coverImg,
		Images:       req.Images,
		VideoURL:     note.VideoURL,
		Content:      note.Content,
		ViewCount:    note.ViewCount,
		LikeCount:    note.LikeCount,
		CollectCount: note.CollectCount,
		CommentCount: note.CommentCount,
		IsLiked:      false,
		IsFavorited:  false,
		Category:     note.Category,
		Tags:         req.Tags,
		Status:       note.Status,
		PublishTime:  note.PublishTime.Format("2006-01-02 15:04:05"),
	}, nil
}
