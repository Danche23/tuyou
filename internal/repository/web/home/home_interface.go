package home

import "tuyou/internal/model/entity"

// HomeRepository 前台首页仓储接口
type HomeRepository interface {
	// ===== 轮播：管理员已标记 is_slide=1 =====
	FindSlideActivities() ([]entity.Activity, error)
	FindSlideNotes() ([]entity.Note, error)
	FindSlideStrategies() ([]entity.Strategy, error)

	// ===== 轮播：补齐候选（is_slide=0，自动选中不写库） =====
	FindFillActivities(limit int) ([]entity.Activity, error)
	FindFillNotes(limit int) ([]entity.Note, error)
	FindFillStrategies(limit int) ([]entity.Strategy, error)

	// ===== 关联数据（批量，避免 N+1） =====
	FindFirstImagesByNoteIDs(noteIDs []uint) (map[uint]string, error)
	FindFirstImagesByStrategyIDs(strategyIDs []uint) (map[uint]string, error)
	FindFirstTagsByNoteIDs(noteIDs []uint) (map[uint]string, error)
	FindFirstTagsByStrategyIDs(strategyIDs []uint) (map[uint]string, error)
}
