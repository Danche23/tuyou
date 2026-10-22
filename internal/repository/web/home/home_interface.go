package home

import (
	"context"
	responseHome "tuyou/internal/model/dto/response/web/home"
	"tuyou/internal/model/entity"
)

// HomeRepository 前台首页仓储接口
type HomeRepository interface {
	GetCachedFirstFeed(ctx context.Context, sort string, category, size int) (*responseHome.FeedResponse, bool)
	SetCachedFirstFeed(ctx context.Context, sort string, category, size int, page *responseHome.FeedResponse)
	FindNoteLikeCounts(ids []uint) (map[uint]int64, error)
	FindStrategyLikeCounts(ids []uint) (map[uint]int64, error)
	// FindSlideActivities ===== 轮播：管理员已标记 is_slide=1 =====
	FindSlideActivities() ([]entity.Activity, error)
	FindSlideNotes() ([]entity.Note, error)
	FindSlideStrategies() ([]entity.Strategy, error)

	// FindFillActivities ===== 轮播：补齐候选（is_slide=0，自动选中不写库） =====
	FindFillActivities(limit int) ([]entity.Activity, error)
	FindFillNotes(limit int) ([]entity.Note, error)
	FindFillStrategies(limit int) ([]entity.Strategy, error)

	// FindFirstImagesByNoteIDs ===== 关联数据（批量，避免 N+1） =====
	FindFirstImagesByNoteIDs(noteIDs []uint) (map[uint]string, error)
	FindFirstImagesByStrategyIDs(strategyIDs []uint) (map[uint]string, error)
	FindFirstTagsByNoteIDs(noteIDs []uint) (map[uint]string, error)
	FindFirstTagsByStrategyIDs(strategyIDs []uint) (map[uint]string, error)

	// CountFeedNotes 统计符合条件的笔记总数
	CountFeedNotes(category int) (int64, error)

	// CountFeedStrategies 统计符合条件的攻略总数
	CountFeedStrategies(category int) (int64, error)

	// FindHotTagIDs 查询热门标签ID（is_hot=1 且启用未删）
	FindHotTagIDs() ([]uint, error)

	// FindHotNoteIDs 查询带热门标签的笔记ID集合（去重）
	FindHotNoteIDs(hotTagIDs []uint) ([]uint, error)

	// FindHotStrategyIDs 查询带热门标签的攻略ID集合（去重）
	FindHotStrategyIDs(hotTagIDs []uint) ([]uint, error)

	// FindFeedNotes 按排序窗口取笔记（limit = offset+size；hotIDs 非空时热门置顶）
	FindFeedNotes(category int, order string, limit int, hotIDs []uint) ([]entity.Note, error)

	// FindFeedStrategies 按排序窗口取攻略（limit = offset+size；hotIDs 非空时热门置顶）
	FindFeedStrategies(category int, order string, limit int, hotIDs []uint) ([]entity.Strategy, error)

	// FindFeedNotesKeyset 游标分页取笔记（每表最多 Limit 条，代价不随页码增长）
	FindFeedNotesKeyset(q FeedKeysetQuery) ([]entity.Note, error)

	// FindFeedStrategiesKeyset 游标分页取攻略（每表最多 Limit 条，代价不随页码增长）
	FindFeedStrategiesKeyset(q FeedKeysetQuery) ([]entity.Strategy, error)

	// FindMembersByIDs 批量查询会员信息
	FindMembersByIDs(memberIDs []uint) (map[uint]*entity.Member, error)

	// FindNoteLikeStatuses 当前用户对各笔记的点赞状态
	FindNoteLikeStatuses(noteIDs []uint, memberID uint) (map[uint]bool, error)

	// FindStrategyLikeStatuses 当前用户对各攻略的点赞状态
	FindStrategyLikeStatuses(strategyIDs []uint, memberID uint) (map[uint]bool, error)
}
