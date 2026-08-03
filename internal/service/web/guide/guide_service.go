package guide

import (
	"context"
	"strconv"
	"time"

	requestGuide "tuyou/internal/model/dto/request/web/guide"
	responseGuide "tuyou/internal/model/dto/response/web/guide"
	"tuyou/internal/model/entity"
	repoGuide "tuyou/internal/repository/web/guide"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"

	"gorm.io/gorm"
)

// 难度映射
var difficultyStrToInt = map[string]int{"入门": 1, "中等": 2, "中高等": 3, "高等": 4}
var difficultyIntToStr = map[int]string{1: "入门", 2: "中等", 3: "中高等", 4: "高等"}

type guideService struct {
	guideRepo repoGuide.GuideRepository
	db        *gorm.DB
}

// NewGuideService 创建攻略服务实例
func NewGuideService(guideRepo repoGuide.GuideRepository, db *gorm.DB) GuideService {
	return &guideService{guideRepo: guideRepo, db: db}
}

// CreateGuide 发布攻略
func (s *guideService) CreateGuide(memberID uint, req *requestGuide.CreateGuideRequest) (*responseGuide.CreateGuideResponse, error) {
	// ========== 输入校验（事务外，快失败） ==========

	// 视频攻略规则：有视频时必须恰好 1 张封面
	if req.VideoURL != "" && len(req.Images) != 1 {
		return nil, errors.New(errors.CodeInvalidParam, "视频攻略只能上传1张封面图片")
	}

	// 难度必须是枚举之一，映射为 int 入库
	difficulty, ok := difficultyStrToInt[req.Difficulty]
	if !ok {
		return nil, errors.New(errors.CodeInvalidParam, "难度不合法")
	}

	tags, err := s.guideRepo.FindTags(s.db, req.Tags)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "标签查询失败", err)
	}
	if len(tags) != len(req.Tags) {
		return nil, errors.New(errors.CodeInvalidParam, "标签不存在或已禁用")
	}

	// ========== 写入（事务内） ==========

	strategy := &entity.Strategy{
		MemberID:    memberID,
		Title:       req.Title,
		Content:     req.Content,
		Category:    req.Category,
		VideoURL:    req.VideoURL,
		Status:      1,
		PublishTime: time.Now(),
		Distance:    req.Distance,
		Duration:    req.Duration,
		Difficulty:  difficulty,
		Transport:   req.Transport,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.guideRepo.Create(tx, strategy); err != nil {
			return err
		}
		tagIDs := make([]uint, len(tags))
		for i, t := range tags {
			tagIDs[i] = t.ID
		}
		if err := s.guideRepo.CreateStrategyTags(tx, strategy.ID, tagIDs); err != nil {
			return err
		}
		if err := s.guideRepo.CreateStrategyImages(tx, strategy.ID, req.Images); err != nil {
			return err
		}
		return s.guideRepo.IncrPublishCount(tx, memberID)
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "发布攻略失败", err)
	}

	// ========== 组装响应（事务外，简洁） ==========

	return &responseGuide.CreateGuideResponse{
		ID:   strconv.FormatUint(uint64(strategy.ID), 10),
		Type: "guide",
	}, nil
}

// GetGuideDetail 获取攻略详情
func (s *guideService) GetGuideDetail(strategyID uint, memberID uint) (*responseGuide.GuideDetailResponse, error) {
	strategy, err := s.guideRepo.FindByID(strategyID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "攻略不存在", err)
	}

	ctx := context.Background()

	// 缓存读取：点赞/收藏状态和点赞数（Redis 优先 → MySQL 回填）
	isLiked := s.readLikeStatus(ctx, strategyID, memberID)
	isFavorited := s.readCollectStatus(ctx, strategyID, memberID)
	likeCount := s.readLikeCount(ctx, strategyID, strategy.LikeCount)

	// 浏览量 +1（非作者，Redis Stream 异步，失败降级直写 DB）
	if strategy.MemberID != memberID {
		if err := s.guideRepo.XAddViewCount(ctx, strategyID); err != nil {
			_ = s.guideRepo.IncrViewCount(strategyID)
		}
		strategy.ViewCount++
	}

	// 作者信息
	member, err := s.guideRepo.GetAuthor(strategy.MemberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}

	// 图片
	noteImages, err := s.guideRepo.FindImagesByStrategyID(strategyID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询图片失败", err)
	}
	images := make([]string, 0, len(noteImages))
	for _, img := range noteImages {
		images = append(images, img.URL)
	}

	// 标签
	tags, err := s.guideRepo.FindTagsByStrategyID(strategyID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询标签失败", err)
	}
	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.TagName)
	}

	// 关注状态
	isFollowed := false
	if strategy.MemberID != memberID {
		isFollowed, err = s.guideRepo.ExistsFollow(memberID, strategy.MemberID)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "查询关注状态失败", err)
		}
	}

	// 封面
	coverImg := ""
	if len(images) > 0 {
		coverImg = images[0]
	}

	// 难度 int → string
	difficulty := difficultyIntToStr[strategy.Difficulty]

	return &responseGuide.GuideDetailResponse{
		ID:           strconv.FormatUint(uint64(strategy.ID), 10),
		Type:         "guide",
		Author:       &responseGuide.AuthorInfo{ID: strconv.FormatUint(uint64(member.ID), 10), Nickname: member.Nickname, Avatar: member.Avatar},
		IsFollowed:   isFollowed,
		Title:        strategy.Title,
		CoverImg:     coverImg,
		Images:       images,
		VideoURL:     strategy.VideoURL,
		Content:      strategy.Content,
		ViewCount:    strategy.ViewCount,
		LikeCount:    likeCount,
		CollectCount: strategy.CollectCount,
		CommentCount: strategy.CommentCount,
		IsLiked:      isLiked,
		IsFavorited:  isFavorited,
		Category:     strategy.Category,
		Tags:         tagNames,
		PublishTime:  strategy.PublishTime.Format("2006-01-02 15:04:05"),
		RouteInfo: &responseGuide.RouteInfo{
			Distance:   strategy.Distance,
			Duration:   strategy.Duration,
			Difficulty: difficulty,
			Transport:  strategy.Transport,
		},
	}, nil
}

// readLikeStatus 读缓存（Redis → MySQL → 回填）
func (s *guideService) readLikeStatus(ctx context.Context, strategyID, memberID uint) bool {
	isLiked, found, err := s.guideRepo.GetCachedLikeStatus(ctx, strategyID, memberID)
	if err == nil && found {
		return isLiked
	}

	isLiked, err = s.guideRepo.ExistsLike(strategyID, memberID)
	if err != nil {
		return false
	}
	s.guideRepo.SetCachedLikeStatus(ctx, strategyID, memberID, isLiked)
	return isLiked
}

func (s *guideService) readLikeCount(ctx context.Context, strategyID uint, fallback int64) int64 {
	count, err := s.guideRepo.GetCachedLikeCount(ctx, strategyID)
	if err == nil {
		return count
	}
	s.guideRepo.SetCachedLikeCount(ctx, strategyID, fallback)
	return fallback
}

// ToggleLikeGuide 点赞/取消点赞（写：MySQL 为准 → 事务 → 同步 Redis）
func (s *guideService) ToggleLikeGuide(strategyID uint, memberID uint) (*responseGuide.LikeToggleResponse, error) {
	ctx := context.Background()

	strategy, err := s.guideRepo.FindByID(strategyID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "攻略不存在", err)
	}

	liked, err := s.guideRepo.ExistsLike(strategyID, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询点赞状态失败", err)
	}

	likeCount := strategy.LikeCount
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if liked {
			if err := s.guideRepo.DeleteLike(tx, strategyID, memberID); err != nil {
				return err
			}
			if err := s.guideRepo.DecrLikeCount(tx, strategyID); err != nil {
				return err
			}
			likeCount--
			return nil
		}
		if err := s.guideRepo.CreateLike(tx, strategyID, memberID); err != nil {
			return err
		}
		if err := s.guideRepo.IncrLikeCount(tx, strategyID); err != nil {
			return err
		}
		likeCount++
		return nil
	})
	if err != nil {
		logger.Errorf("点赞事务失败 strategyID=%d memberID=%d err=%v", strategyID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "点赞操作失败", err)
	}

	// 同步缓存
	s.guideRepo.SetCachedLikeStatus(ctx, strategyID, memberID, !liked)
	s.guideRepo.SetCachedLikeCount(ctx, strategyID, likeCount)

	return &responseGuide.LikeToggleResponse{IsLiked: !liked, LikeCount: likeCount}, nil
}

// readCollectStatus 读缓存（Redis → MySQL → 回填）
func (s *guideService) readCollectStatus(ctx context.Context, strategyID, memberID uint) bool {
	isFavorited, found, err := s.guideRepo.GetCachedCollectStatus(ctx, strategyID, memberID)
	if err == nil && found {
		return isFavorited
	}

	isFavorited, err = s.guideRepo.ExistsCollect(strategyID, memberID)
	if err != nil {
		return false
	}
	s.guideRepo.SetCachedCollectStatus(ctx, strategyID, memberID, isFavorited)
	return isFavorited
}

// ToggleCollectGuide 收藏/取消收藏（写：MySQL 为准 → 事务 → 同步 Redis）
func (s *guideService) ToggleCollectGuide(strategyID uint, memberID uint) (*responseGuide.CollectToggleResponse, error) {
	ctx := context.Background()

	_, err := s.guideRepo.FindByID(strategyID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "攻略不存在", err)
	}

	favorited, err := s.guideRepo.ExistsCollect(strategyID, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询收藏状态失败", err)
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if favorited {
			if err := s.guideRepo.DeleteCollect(tx, strategyID, memberID); err != nil {
				return err
			}
			return s.guideRepo.DecrCollectCount(tx, strategyID)
		}
		if err := s.guideRepo.CreateCollect(tx, strategyID, memberID); err != nil {
			return err
		}
		return s.guideRepo.IncrCollectCount(tx, strategyID)
	})
	if err != nil {
		logger.Errorf("收藏事务失败 strategyID=%d memberID=%d err=%v", strategyID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "收藏操作失败", err)
	}

	// 同步缓存
	s.guideRepo.SetCachedCollectStatus(ctx, strategyID, memberID, !favorited)

	return &responseGuide.CollectToggleResponse{IsFavorited: !favorited}, nil
}
