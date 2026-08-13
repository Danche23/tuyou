package search

import "tuyou/internal/model/entity"

// SearchRepository 前台全量搜索仓储接口
type SearchRepository interface {
	// FindHotSearchNames 查询热门搜索关键词（status=1 AND is_hot=1 且未删除），按 id 升序
	FindHotSearchNames() ([]string, error)

	// FindTagIDsByKeyword 关键词命中标签名（status=1 且未删除），返回标签ID集合
	FindTagIDsByKeyword(keyword string) ([]uint, error)

	// FindNoteIDsByTagIDs 带指定标签的笔记ID集合（去重）
	FindNoteIDsByTagIDs(tagIDs []uint) ([]uint, error)

	// FindStrategyIDsByTagIDs 带指定标签的攻略ID集合（去重）
	FindStrategyIDsByTagIDs(tagIDs []uint) ([]uint, error)

	// SearchNotes 搜索笔记（标题/分类/标签 OR 命中），按热度倒序，取 limit 条（limit=offset+size）
	// contentIDs 为关键词命中的标签所对应的笔记ID集合（可为空）
	SearchNotes(keyword string, category int, contentIDs []uint, limit int) ([]entity.Note, error)

	// CountSearchNotes 统计命中笔记数
	CountSearchNotes(keyword string, category int, contentIDs []uint) (int64, error)

	// SearchStrategies 搜索攻略（标题/分类/标签 OR 命中），按热度倒序，取 limit 条（limit=offset+size）
	// contentIDs 为关键词命中的标签所对应的攻略ID集合（可为空）
	SearchStrategies(keyword string, category int, contentIDs []uint, limit int) ([]entity.Strategy, error)

	// CountSearchStrategies 统计命中攻略数
	CountSearchStrategies(keyword string, category int, contentIDs []uint) (int64, error)

	// SearchActivities 搜索活动（标题/分类/地点 OR 命中），按开始时间倒序，分页（与前台活动列表排序一致）
	SearchActivities(keyword string, category int, offset, size int) ([]entity.Activity, error)

	// CountSearchActivities 统计命中活动数
	CountSearchActivities(keyword string, category int) (int64, error)
}
