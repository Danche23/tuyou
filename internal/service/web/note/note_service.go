package note

import (
	"context"
	"fmt"
	"strconv"
	"time"

	requestNote "tuyou/internal/model/dto/request/web/note"
	responseNote "tuyou/internal/model/dto/response/web/note"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	repoNote "tuyou/internal/repository/web/note"
	"tuyou/internal/stream"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"

	"gorm.io/gorm"
)

type noteService struct {
	noteRepo  repoNote.NoteRepository
	db        *gorm.DB
	redisRepo repository.RedisRepository
}

// NewNoteService 创建笔记服务实例
func NewNoteService(noteRepo repoNote.NoteRepository, db *gorm.DB, redisRepo repository.RedisRepository) NoteService {
	return &noteService{noteRepo: noteRepo, db: db, redisRepo: redisRepo}
}

// CreateNote 发布笔记
func (s *noteService) CreateNote(memberID uint, req *requestNote.CreateNoteRequest) (*responseNote.CreateNoteResponse, error) {
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

	member, err := s.noteRepo.GetAuthor(memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}

	authorInfo := &responseNote.AuthorInfo{
		ID:       strconv.FormatUint(uint64(member.ID), 10),
		Nickname: member.Nickname,
		Avatar:   member.Avatar,
	}

	coverImg := ""
	if len(req.Images) > 0 {
		coverImg = req.Images[0]
	}
	return &responseNote.CreateNoteResponse{
		ID:           strconv.FormatUint(uint64(note.ID), 10),
		Type:         "note",
		Author:       authorInfo,
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

	// 1. 查笔记主表
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}

	// 2. 非作者本人，浏览量 +1（优先 Redis Stream，失败降级直写 DB）
	if note.MemberID != memberID {
		msg := map[string]interface{}{
			"type": "note",
			"id":   noteID,
		}
		if err := s.redisRepo.XAdd(ctx, stream.StreamKey, msg); err != nil {
			// 降级：Redis 不可用 → 直写 DB
			_ = s.noteRepo.IncrViewCount(noteID)
		}
		note.ViewCount++ // 当前响应立即 +1
	}

	// 3. 查作者
	member, err := s.noteRepo.GetAuthor(note.MemberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}
	authorInfo := &responseNote.AuthorInfo{
		ID:       strconv.FormatUint(uint64(member.ID), 10),
		Nickname: member.Nickname,
		Avatar:   member.Avatar,
	}

	// 4. 查图片
	noteImages, err := s.noteRepo.FindImagesByNoteID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询图片失败", err)
	}
	images := make([]string, 0, len(noteImages))
	for _, img := range noteImages {
		images = append(images, img.URL)
	}

	// 5. 查标签
	tags, err := s.noteRepo.FindTagsByNoteID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}
	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.TagName)
	}

	// 6. 是否已点赞（Redis 优先，失败查 MySQL 并回填）
	key := fmt.Sprintf("note:likes:%d", noteID)
	var isLiked bool
	isLiked, err = s.redisRepo.SIsMember(ctx, key, memberID)
	if err != nil {
		isLiked, err = s.noteRepo.ExistsLike(noteID, memberID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询点赞状态失败", err)
		}
		// 回填 Redis
		if isLiked {
			_ = s.redisRepo.SAdd(ctx, key, memberID)
		}
	}

	// 7. 是否已收藏（Redis 优先，失败查 MySQL 并回填）
	collectKey := fmt.Sprintf("note:collects:%d", noteID)
	var isFavorited bool
	isFavorited, err = s.redisRepo.SIsMember(ctx, collectKey, memberID)
	if err != nil {
		isFavorited, err = s.noteRepo.ExistsCollect(noteID, memberID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询收藏状态失败", err)
		}
		// 回填 Redis
		if isFavorited {
			_ = s.redisRepo.SAdd(ctx, collectKey, memberID)
		}
	}

	// 8. 是否关注（自己不能关注自己）
	isFollowed := false
	if note.MemberID != memberID {
		isFollowed, err = s.noteRepo.ExistsFollow(memberID, note.MemberID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询关注状态失败", err)
		}
	}

	// 9. 封面图
	coverImg := ""
	if len(images) > 0 {
		coverImg = images[0]
	}

	// 10. 组装响应
	return &responseNote.NoteDetailResponse{
		ID:           strconv.FormatUint(uint64(note.ID), 10),
		Type:         "note",
		Author:       authorInfo,
		IsFollowed:   isFollowed,
		Title:        note.Title,
		CoverImg:     coverImg,
		Images:       images,
		VideoURL:     note.VideoURL,
		Content:      note.Content,
		ViewCount:    note.ViewCount,
		LikeCount:    note.LikeCount,
		CollectCount: note.CollectCount,
		CommentCount: note.CommentCount,
		IsLiked:      isLiked,
		IsFavorited:  isFavorited,
		Category:     note.Category,
		Tags:         tagNames,
		Status:       note.Status,
		PublishTime:  note.PublishTime.Format("2006-01-02 15:04:05"),
	}, nil
}

// ToggleLikeNote 点赞/取消点赞（MySQL 最终存储 + Redis 缓存）
func (s *noteService) ToggleLikeNote(noteID uint, memberID uint) (*responseNote.LikeToggleResponse, error) {
	ctx := context.Background()
	// 查询笔记
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}
	key := fmt.Sprintf("note:likes:%d", noteID)
	// Redis 判断点赞状态
	liked, err := s.redisRepo.SIsMember(ctx, key, memberID)

	if err != nil {
		// Redis异常，降级MySQL
		liked, err = s.noteRepo.ExistsLike(noteID, memberID)

		if err != nil {
			return nil, errors.NewWithErr(
				errors.CodeInternalError,
				"查询点赞状态失败",
				err,
			)
		}
		// 回填Redis
		if liked {
			_ = s.redisRepo.SAdd(ctx, key, memberID)
		}
	}
	// 当前点赞数量
	likeCount := note.LikeCount
	// MySQL事务（note_like + like_count）
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

	// 同步 Redis 缓存
	if liked {
		if err := s.redisRepo.SRem(ctx, key, memberID); err != nil {
			logger.Errorf("同步 Redis 缓存失败（取消点赞）: %v", err)
		}
	} else {
		if err := s.redisRepo.SAdd(ctx, key, memberID); err != nil {
			logger.Errorf("同步 Redis 缓存失败（点赞）: %v", err)
		}
	}

	return &responseNote.LikeToggleResponse{
		IsLiked:   !liked,
		LikeCount: likeCount,
	}, nil
}

// ToggleCollectNote 收藏/取消收藏（MySQL 最终存储 + Redis 缓存）
func (s *noteService) ToggleCollectNote(noteID uint, memberID uint) (*responseNote.CollectToggleResponse, error) {
	ctx := context.Background()

	// 查询笔记是否存在
	_, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}

	key := fmt.Sprintf("note:collects:%d", noteID)

	// Redis 判断收藏状态
	favorite, err := s.redisRepo.SIsMember(ctx, key, memberID)
	if err != nil {
		// Redis异常，降级MySQL
		favorite, err = s.noteRepo.ExistsCollect(noteID, memberID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询收藏状态失败", err)
		}
		// 回填Redis
		if favorite {
			_ = s.redisRepo.SAdd(ctx, key, memberID)
		}
	}

	// MySQL事务（note_collect + collect_count）
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if favorite {
			if err := s.noteRepo.DeleteCollect(tx, noteID, memberID); err != nil {
				return err
			}
			return s.noteRepo.DecrCollectCount(tx, noteID)
		}
		if err := s.noteRepo.CreateCollect(tx, noteID, memberID); err != nil {
			return err
		}
		return s.noteRepo.IncrCollectCount(tx, noteID)
	})
	if err != nil {
		logger.Errorf("收藏事务失败 noteID=%d memberID=%d err=%v", noteID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "收藏操作失败", err)
	}

	// 同步 Redis 缓存
	if favorite {
		if err := s.redisRepo.SRem(ctx, key, memberID); err != nil {
			logger.Errorf("同步 Redis 缓存失败（取消收藏）: %v", err)
		}
	} else {
		if err := s.redisRepo.SAdd(ctx, key, memberID); err != nil {
			logger.Errorf("同步 Redis 缓存失败（收藏）: %v", err)
		}
	}

	return &responseNote.CollectToggleResponse{
		IsFavorited: !favorite,
	}, nil
}

// DeleteNote 删除笔记（下架，仅作者本人）
func (s *noteService) DeleteNote(noteID uint, memberID uint) error {
	// 1. 查笔记（FindByID 只查 status=1，已删除则 404）
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return errors.NewWithErr(errors.CodeNotFound, "笔记不存在", err)
	}

	// 2. 仅作者本人可删除
	if note.MemberID != memberID {
		return errors.New(errors.CodeForbidden, "无权删除他人笔记")
	}

	// 3. 事务：下架笔记 + 下架评论 + 发布数-1
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
