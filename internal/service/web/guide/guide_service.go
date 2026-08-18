package guide

import (
	"context"
	"strconv"
	"time"

	requestGuide "tuyou/internal/model/dto/request/web/guide"
	responseGuide "tuyou/internal/model/dto/response/web/guide"
	"tuyou/internal/model/entity"
	repoGuide "tuyou/internal/repository/web/guide"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/errors"
	"tuyou/pkg/logger"
	"tuyou/pkg/sensitive"

	"gorm.io/gorm"
)

// 难度映射
var difficultyStrToInt = map[string]int{"入门": 1, "中等": 2, "中高等": 3, "高等": 4}
var difficultyIntToStr = map[int]string{1: "入门", 2: "中等", 3: "中高等", 4: "高等"}

type guideService struct {
	guideRepo repoGuide.GuideRepository
	db        *gorm.DB
	sensitive *sensitive.SensitiveMatcher
	notifier  *serviceMsg.Notifier
}

// NewGuideService 创建攻略服务实例
func NewGuideService(guideRepo repoGuide.GuideRepository, db *gorm.DB, matcher *sensitive.SensitiveMatcher, notifier *serviceMsg.Notifier) GuideService {
	return &guideService{guideRepo: guideRepo, db: db, sensitive: matcher, notifier: notifier}
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

	// 敏感词快照：命中置 is_sensitive=1，内容存原文（设计 §8.5）
	isSensitive := 0
	if s.sensitive.IsSensitive(req.Title, req.Content, req.Transport) {
		isSensitive = 1
	}

	strategy := &entity.Strategy{
		MemberID:    memberID,
		Title:       req.Title,
		Content:     req.Content,
		Category:    req.Category,
		VideoURL:    req.VideoURL,
		Status:      1,
		Heat:        10, // 初始基础热度，与 heat.go 公式中的基础热度10保持一致
		PublishTime: time.Now(),
		Distance:    req.Distance,
		Duration:    req.Duration,
		Difficulty:  difficulty,
		Transport:   req.Transport,
		IsSensitive: isSensitive,
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
		Author:       &responseGuide.AuthorInfo{ID: strconv.FormatUint(uint64(member.ID), 10), Nickname: s.mask(member.Nickname), Avatar: member.Avatar},
		IsFollowed:   isFollowed,
		IsOwner:      strategy.MemberID == memberID,
		Title:        s.mask(strategy.Title),
		CoverImg:     coverImg,
		Images:       images,
		VideoURL:     strategy.VideoURL,
		Content:      s.mask(strategy.Content),
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
			Transport:  s.mask(strategy.Transport),
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

	// 事务：仓库方法内部已按「实际插入/删除是否成功」决定是否增减计数，天然幂等
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if liked {
			return s.guideRepo.DeleteLike(tx, strategyID, memberID)
		}
		return s.guideRepo.CreateLike(tx, strategyID, memberID)
	})
	if err != nil {
		logger.Errorf("点赞事务失败 strategyID=%d memberID=%d err=%v", strategyID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "点赞操作失败", err)
	}

	// 同步缓存：状态写本次结果；数量用事务后重查的真实值，避免并发下「旧值±1」写脏缓存
	// 点赞成功 → 给内容作者发实时通知（取消赞不通知；自己赞自己不通知）
	if !liked {
		s.notifier.NotifyLike(strategy.MemberID, memberID, strategy.Title)
	}

	// 同步缓存
	s.guideRepo.SetCachedLikeStatus(ctx, strategyID, memberID, !liked)
	if fresh, err := s.guideRepo.FindByID(strategyID); err == nil {
		strategy = fresh
	}
	s.guideRepo.SetCachedLikeCount(ctx, strategyID, strategy.LikeCount)

	return &responseGuide.LikeToggleResponse{IsLiked: !liked, LikeCount: strategy.LikeCount}, nil
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

	strategy, err := s.guideRepo.FindByID(strategyID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "攻略不存在", err)
	}

	favorited, err := s.guideRepo.ExistsCollect(strategyID, memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "查询收藏状态失败", err)
	}

	// 事务：仓库方法内部已按「实际插入/删除是否成功」决定是否增减计数，天然幂等
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if favorited {
			return s.guideRepo.DeleteCollect(tx, strategyID, memberID)
		}
		return s.guideRepo.CreateCollect(tx, strategyID, memberID)
	})
	if err != nil {
		logger.Errorf("收藏事务失败 strategyID=%d memberID=%d err=%v", strategyID, memberID, err)
		return nil, errors.NewWithErr(errors.CodeInternalError, "收藏操作失败", err)
	}

	// 同步缓存（攻略收藏数无 Redis 缓存，只更新状态）
	// 收藏成功 → 给内容作者发实时通知（取消收藏不通知）
	if !favorited {
		s.notifier.NotifyCollect(strategy.MemberID, memberID, strategy.Title)
	}

	// 同步缓存
	s.guideRepo.SetCachedCollectStatus(ctx, strategyID, memberID, !favorited)

	return &responseGuide.CollectToggleResponse{IsFavorited: !favorited}, nil
}

// DeleteGuide 删除攻略（下架，仅作者本人）
func (s *guideService) DeleteGuide(strategyID uint, memberID uint) error {
	strategy, err := s.guideRepo.FindByID(strategyID)
	if err != nil {
		return errors.NewWithErr(errors.CodeNotFound, "攻略不存在", err)
	}
	if strategy.MemberID != memberID {
		return errors.New(errors.CodeForbidden, "无权删除他人攻略")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		changed, err := s.guideRepo.UpdateStatus(tx, strategyID, 0)
		if err != nil {
			return err
		}
		if !changed {
			return nil // 已被并发删除/下架，幂等跳过，不再扣发布数
		}
		if err := s.guideRepo.SoftDelete(tx, strategyID); err != nil {
			return err
		}
		if err := s.guideRepo.UpdateCommentStatusByTarget(tx, 2, strategyID, 0); err != nil {
			return err
		}
		return s.guideRepo.DecrPublishCount(tx, memberID)
	})
	if err != nil {
		logger.Errorf("删除攻略事务失败 strategyID=%d memberID=%d err=%v", strategyID, memberID, err)
		return errors.NewWithErr(errors.CodeInternalError, "删除攻略失败", err)
	}

	return nil
}

// mask 前台文本出口统一打星（nil 安全）
func (s *guideService) mask(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}
