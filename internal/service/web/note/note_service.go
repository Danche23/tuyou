package note

import (
	"context"
	"strconv"
	"time"

	requestNote "tuyou/internal/model/dto/request/web/note"
	responseNote "tuyou/internal/model/dto/response/web/note"
	"tuyou/internal/model/entity"
	repoNote "tuyou/internal/repository/web/note"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"
	"tuyou/pkg/sensitive"

	"gorm.io/gorm"
)

type noteService struct {
	noteRepo  repoNote.NoteRepository
	db        *gorm.DB
	sensitive *sensitive.SensitiveMatcher
	notifier  *serviceMsg.Notifier
}

// NewNoteService 创建笔记服务实例
func NewNoteService(noteRepo repoNote.NoteRepository, db *gorm.DB, matcher *sensitive.SensitiveMatcher, notifier *serviceMsg.Notifier) NoteService {
	return &noteService{noteRepo: noteRepo, db: db, sensitive: matcher, notifier: notifier}
}

// CreateNote 发布笔记
func (s *noteService) CreateNote(memberID uint, req *requestNote.CreateNoteRequest) (*responseNote.CreateNoteResponse, error) {
	// ========== 输入校验（事务外，快失败） ==========

	if req.VideoURL != "" && len(req.Images) != 1 {
		return nil, errors.New(errors.CodeInvalidParam, "视频笔记只能上传1张封面图片")
	}

	// 图片数量校验（1~5 张）——放 service 层返回具体错误，避免 DTO binding 拦截返回笼统的"参数校验失败"
	if len(req.Images) < 1 || len(req.Images) > 5 {
		return nil, errors.New(errors.CodeInvalidParam, "图片数量需为1~5张")
	}

	tags, err := s.noteRepo.FindTags(s.db, req.Tags)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "标签查询失败", err)
	}
	if len(tags) != len(req.Tags) {
		return nil, errors.New(errors.CodeInvalidParam, "标签不存在或已禁用")
	}

	// ========== 写入（事务内） ==========

	// 敏感词快照：命中置 is_sensitive=1，内容存原文（设计 §8.5）
	isSensitive := 0
	if s.sensitive.IsSensitive(req.Title, req.Content) {
		isSensitive = 1
	}

	note := &entity.Note{
		MemberID:    memberID,
		Title:       req.Title,
		Content:     req.Content,
		Category:    req.Category,
		VideoURL:    req.VideoURL,
		Status:      1,
		Heat:        10, // 初始基础热度，与 heat.go 公式中的基础热度10保持一致
		PublishTime: time.Now(),
		IsSensitive: isSensitive,
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
		Author:       &responseNote.AuthorInfo{ID: strconv.FormatUint(uint64(member.ID), 10), Nickname: s.mask(member.Nickname), Avatar: member.Avatar},
		Title:        s.mask(note.Title),
		CoverImg:     coverImg,
		Images:       req.Images,
		VideoURL:     note.VideoURL,
		Content:      s.mask(note.Content),
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
		Author:       &responseNote.AuthorInfo{ID: strconv.FormatUint(uint64(member.ID), 10), Nickname: s.mask(member.Nickname), Avatar: member.Avatar},
		IsFollowed:   isFollowed,
		IsOwner:      note.MemberID == memberID,
		Title:        s.mask(note.Title),
		CoverImg:     coverImg,
		Images:       images,
		VideoURL:     note.VideoURL,
		Content:      s.mask(note.Content),
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

	// 事务：仓库方法内部已按「实际插入/删除是否成功」决定是否增减计数，天然幂等
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if liked {
			return s.noteRepo.DeleteLike(tx, noteID, memberID)
		}
		return s.noteRepo.CreateLike(tx, noteID, memberID)
	})
	if err != nil {
		logger.Errorf("点赞事务失败 noteID=%d memberID=%d err=%v", noteID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "点赞操作失败", err)
	}

	// 同步缓存：状态写本次结果；数量用事务后重查的真实值，避免并发下「旧值±1」写脏缓存
	// 点赞成功 → 给内容作者发实时通知（取消赞不通知；自己赞自己不通知）
	if !liked {
		s.notifier.NotifyLike(note.MemberID, memberID, note.Title)
	}

	// 同步缓存
	s.noteRepo.SetCachedLikeStatus(ctx, noteID, memberID, !liked)
	if fresh, err := s.noteRepo.FindByID(noteID); err == nil {
		note = fresh
	}
	s.noteRepo.SetCachedLikeCount(ctx, noteID, note.LikeCount)

	return &responseNote.LikeToggleResponse{IsLiked: !liked, LikeCount: note.LikeCount}, nil
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

	// 事务：仓库方法内部已按「实际插入/删除是否成功」决定是否增减计数，天然幂等
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if favorite {
			return s.noteRepo.DeleteCollect(tx, noteID, memberID)
		}
		return s.noteRepo.CreateCollect(tx, noteID, memberID)
	})
	if err != nil {
		logger.Errorf("收藏事务失败 noteID=%d memberID=%d err=%v", noteID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "收藏操作失败", err)
	}

	// 同步缓存：状态写本次结果；数量用事务后重查的真实值，避免并发下「旧值±1」写脏缓存
	// 收藏成功 → 给内容作者发实时通知（取消收藏不通知）
	if !favorite {
		s.notifier.NotifyCollect(note.MemberID, memberID, note.Title)
	}

	s.noteRepo.SetCachedCollectStatus(ctx, noteID, memberID, !favorite)
	if fresh, err := s.noteRepo.FindByID(noteID); err == nil {
		note = fresh
	}
	s.noteRepo.SetCachedCollectCount(ctx, noteID, note.CollectCount)

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
		changed, err := s.noteRepo.UpdateStatus(tx, noteID, 0)
		if err != nil {
			return err
		}
		if !changed {
			return nil // 已被并发删除/下架，幂等跳过，不再扣发布数
		}
		if err := s.noteRepo.SoftDelete(tx, noteID); err != nil {
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

// mask 前台文本出口统一打星（nil 安全）
func (s *noteService) mask(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}
