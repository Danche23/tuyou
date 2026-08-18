package search

import "tuyou/internal/model/entity"

// SearchRepository 前台全量搜索仓储接口
type SearchRepository interface {
	// FindHotSearchNames 查询热门搜索关键词（status=1 AND is_hot=1 且未删除），按 id 升序
	FindHotSearchNames() ([]string, error)

	// UpsertSearchHistory 记录一条搜索历史（同词去重置顶，每人限最近 HistoryLimit 条）
	UpsertSearchHistory(memberID uint, keyword string) error

	// ListSearchHistory 查询某用户最近搜索历史（create_time 倒序，最多 limit 条）
	ListSearchHistory(memberID uint, limit int) ([]entity.SearchHistory, error)

	// DeleteSearchHistoryByID 删除单条搜索历史（硬删，仅本人，幂等）
	DeleteSearchHistoryByID(memberID, id uint) error

	// ClearSearchHistory 清空某用户全部搜索历史（硬删，幂等）
	ClearSearchHistory(memberID uint) error

	// SearchNotes 搜索笔记（标题/正文/分类 OR 命中），相关度优先、热度兜底，取 limit 条（limit=offset+size）
	SearchNotes(keyword string, category int, limit int) ([]entity.Note, error)

	// CountSearchNotes 统计命中笔记数
	CountSearchNotes(keyword string, category int) (int64, error)

	// SearchStrategies 搜索攻略（标题/正文/分类 OR 命中），相关度优先、热度兜底，取 limit 条（limit=offset+size）
	SearchStrategies(keyword string, category int, limit int) ([]entity.Strategy, error)

	// CountSearchStrategies 统计命中攻略数
	CountSearchStrategies(keyword string, category int) (int64, error)

	// SearchActivities 搜索活动（标题/分类/地点 OR 命中），按开始时间倒序，分页（与前台活动列表排序一致）
	SearchActivities(keyword string, category int, offset, size int) ([]entity.Activity, error)

	// CountSearchActivities 统计命中活动数
	CountSearchActivities(keyword string, category int) (int64, error)
}
