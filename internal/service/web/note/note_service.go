package note

import (
	"context"
	"strconv"
	"time"

	requestNote "tuyou/internal/model/dto/request/web/note"
	responseNote "tuyou/internal/model/dto/response/web/note"
	"tuyou/internal/model/entity"
	repoNote "tuyou/internal/repository/web/note"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"

	"gorm.io/gorm"
)

type noteService struct {
	noteRepo repoNote.NoteRepository
	db       *gorm.DB
}

// NewNoteService 创建笔记服务实例
func NewNoteService(noteRepo repoNote.NoteRepository, db *gorm.DB) NoteService {
	return &noteService{noteRepo: noteRepo, db: db}
}

// CreateNote 发布笔记
func (s *noteService) CreateNote(memberID uint, req *requestNote.CreateNoteRequest) (*responseNote.CreateNoteResponse, error) {
	// ========== 输入校验（事务外，快失败） ==========

	if req.VideoURL != "" && len(req.Images) != 1 {
		return nil, errors.New(errors.CodeInvalidParam, "视频笔记只能上传1张封面图片")
	}

	tags, err := s.noteRepo.FindTags(s.db, req.Tags)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "标签查询失败", err)
	}
	if len(tags) != len(req.Tags) {
		return nil, errors.New(errors.CodeInvalidParam, "标签不存在或已禁用")
	}

	// ========== 写入（事务内） ==========

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
		if err := s.noteRepo.Create(tx, note); err != nil {
			return err
		}
		tagIDs := make([]uint, len(tags))
		for i, t := range tags {
			tagIDs[i] = t.ID
		}
		if err := s.noteRepo.CreateNoteTags(tx, note.ID, tagIDs); err != nil {
			return err
		}
		if err := s.noteRepo.CreateNoteImages(tx, note.ID, req.Images); err != nil {
			return err
		}
		return s.noteRepo.IncrPublishCount(tx, memberID)
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "发布笔记失败", err)
	}

	// ========== 组装响应（事务外） ==========

	member, err := s.noteRepo.GetAuthor(memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}

	coverImg := ""
	if len(req.Images) > 0 {
		coverImg = req.Images[0]
	}
	return &responseNote.CreateNoteResponse{
		ID:           strconv.FormatUint(uint64(note.ID), 10),
		Type:         "note",
		Author:       &responseNote.AuthorInfo{ID: strconv.FormatUint(uint64(member.ID), 10), Nickname: member.Nickname, Avatar: member.Avatar},
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

// GetNoteDetail 获取笔记详情
func (s *noteService) GetNoteDetail(noteID uint, memberID uint) (*responseNote.NoteDetailResponse, error) {
	ctx := context.Background()

	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}

	// 浏览量 +1（非作者，Redis Stream 异步，失败降级直写 DB）
	if note.MemberID != memberID {
		if err := s.noteRepo.XAddViewCount(ctx, noteID); err != nil {
			_ = s.noteRepo.IncrViewCount(noteID)
		}
		note.ViewCount++
	}

	member, err := s.noteRepo.GetAuthor(note.MemberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}

	noteImages, err := s.noteRepo.FindImagesByNoteID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询图片失败", err)
	}
	images := make([]string, 0, len(noteImages))
	for _, img := range noteImages {
		images = append(images, img.URL)
	}

	tags, err := s.noteRepo.FindTagsByNoteID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}
	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.TagName)
	}

	// 缓存读取：点赞/收藏的状态和数量（Redis 优先 → MySQL 回填）
	isLiked := s.readLikeStatus(ctx, noteID, memberID)

	isFavorited := s.readCollectStatus(ctx, noteID, memberID)

	likeCount := s.readLikeCount(ctx, noteID, note.LikeCount)

	collectCount := s.readCollectCount(ctx, noteID, note.CollectCount)

	isFollowed := false
	if note.MemberID != memberID {
		isFollowed, err = s.noteRepo.ExistsFollow(memberID, note.MemberID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询关注状态失败", err)
		}
	}

	coverImg := ""
	if len(images) > 0 {
		coverImg = images[0]
	}

	return &responseNote.NoteDetailResponse{
		ID:           strconv.FormatUint(uint64(note.ID), 10),
		Type:         "note",
		Author:       &responseNote.AuthorInfo{ID: strconv.FormatUint(uint64(member.ID), 10), Nickname: member.Nickname, Avatar: member.Avatar},
		IsFollowed:   isFollowed,
		Title:        note.Title,
		CoverImg:     coverImg,
		Images:       images,
		VideoURL:     note.VideoURL,
		Content:      note.Content,
		ViewCount:    note.ViewCount,
		LikeCount:    likeCount,
		CollectCount: collectCount,
		CommentCount: note.CommentCount,
		IsLiked:      isLiked,
		IsFavorited:  isFavorited,
		Category:     note.Category,
		Tags:         tagNames,
		Status:       note.Status,
		PublishTime:  note.PublishTime.Format("2006-01-02 15:04:05"),
	}, nil
}

// ToggleLikeNote 点赞/取消点赞（写：MySQL 为准 → 事务 → 同步 Redis）
func (s *noteService) ToggleLikeNote(noteID uint, memberID uint) (*responseNote.LikeToggleResponse, error) {
	ctx := context.Background()

	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}

	liked, err := s.noteRepo.ExistsLike(noteID, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询点赞状态失败", err)
	}

	likeCount := note.LikeCount
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if liked {
			if err := s.noteRepo.DeleteLike(tx, noteID, memberID); err != nil {
				return err
			}
			if err := s.noteRepo.DecrLikeCount(tx, noteID); err != nil {
				return err
			}
			likeCount--
			return nil
		}
		if err := s.noteRepo.CreateLike(tx, noteID, memberID); err != nil {
			return err
		}
		if err := s.noteRepo.IncrLikeCount(tx, noteID); err != nil {
			return err
		}
		likeCount++
		return nil
	})
	if err != nil {
		logger.Errorf("点赞事务失败 noteID=%d memberID=%d err=%v", noteID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "点赞操作失败", err)
	}

	// 同步缓存
	s.noteRepo.SetCachedLikeStatus(ctx, noteID, memberID, !liked)
	s.noteRepo.SetCachedLikeCount(ctx, noteID, likeCount)

	return &responseNote.LikeToggleResponse{IsLiked: !liked, LikeCount: likeCount}, nil
}

// ToggleCollectNote 收藏/取消收藏（写：MySQL 为准 → 事务 → 同步 Redis）
func (s *noteService) ToggleCollectNote(noteID uint, memberID uint) (*responseNote.CollectToggleResponse, error) {
	ctx := context.Background()

	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}

	favorite, err := s.noteRepo.ExistsCollect(noteID, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询收藏状态失败", err)
	}

	collectCount := note.CollectCount
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if favorite {
			if err := s.noteRepo.DeleteCollect(tx, noteID, memberID); err != nil {
				return err
			}
			if err := s.noteRepo.DecrCollectCount(tx, noteID); err != nil {
				return err
			}
			collectCount--
			return nil
		}
		if err := s.noteRepo.CreateCollect(tx, noteID, memberID); err != nil {
			return err
		}
		if err := s.noteRepo.IncrCollectCount(tx, noteID); err != nil {
			return err
		}
		collectCount++
		return nil
	})
	if err != nil {
		logger.Errorf("收藏事务失败 noteID=%d memberID=%d err=%v", noteID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "收藏操作失败", err)
	}

	s.noteRepo.SetCachedCollectStatus(ctx, noteID, memberID, !favorite)

	s.noteRepo.SetCachedCollectCount(ctx, noteID, collectCount)

	return &responseNote.CollectToggleResponse{IsFavorited: !favorite}, nil
}

// DeleteNote 删除笔记（下架，仅作者本人）
func (s *noteService) DeleteNote(noteID uint, memberID uint) error {
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}
	if note.MemberID != memberID {
		return errors.New(errors.CodeForbidden, "无权删除他人笔记")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.noteRepo.UpdateStatus(tx, noteID, 0); err != nil {
			return err
		}
		if err := s.noteRepo.UpdateCommentStatusByTarget(tx, 1, noteID, 0); err != nil {
			return err
		}
		return s.noteRepo.DecrPublishCount(tx, memberID)
	})
	if err != nil {
		logger.Errorf("删除笔记事务失败 noteID=%d memberID=%d err=%v", noteID, memberID, err)
		return errors.NewWithErr(errors.CodeInternalError, "删除笔记失败", err)
	}

	return nil
}

// readLikeStatus 读缓存（Redis → MySQL → 回填）

func (s *noteService) readLikeStatus(ctx context.Context, noteID, memberID uint) bool {
	isLiked, found, err := s.noteRepo.GetCachedLikeStatus(ctx, noteID, memberID)
	if err == nil && found {
		return isLiked
	}

	isLiked, err = s.noteRepo.ExistsLike(noteID, memberID)
	if err != nil {
		return false
	}
	s.noteRepo.SetCachedLikeStatus(ctx, noteID, memberID, isLiked)
	return isLiked
}

func (s *noteService) readCollectStatus(ctx context.Context, noteID, memberID uint) bool {
	isFavorited, found, err := s.noteRepo.GetCachedCollectStatus(ctx, noteID, memberID)
	if err == nil && found {
		return isFavorited
	}

	isFavorited, err = s.noteRepo.ExistsCollect(noteID, memberID)
	if err != nil {
		return false
	}
	s.noteRepo.SetCachedCollectStatus(ctx, noteID, memberID, isFavorited)
	return isFavorited
}

func (s *noteService) readLikeCount(ctx context.Context, noteID uint, fallback int64) int64 {
	count, err := s.noteRepo.GetCachedLikeCount(ctx, noteID)
	if err == nil {
		return count
	}
	s.noteRepo.SetCachedLikeCount(ctx, noteID, fallback)
	return fallback
}

func (s *noteService) readCollectCount(ctx context.Context, noteID uint, fallback int64) int64 {
	count, err := s.noteRepo.GetCachedCollectCount(ctx, noteID)
	if err == nil {
		return count
	}
	s.noteRepo.SetCachedCollectCount(ctx, noteID, fallback)
	return fallback
}
