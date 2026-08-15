package task

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	repoComment "tuyou/internal/repository/web/comment"
	"tuyou/pkg/logger"
)

// StartCommentHeatJob 启动评论热度与 comment_count 校准定时任务：
// 每小时执行一次，启动时先跑一次。
// ① comment_count 校准：把 note/strategy 两表的评论总数用 COUNT 从 comment 表重算（软删过滤），
//
//	兜底实时 ±1 可能漏掉的路径（如后台硬删笔记、手动改库等）。
//
// ② 评论热度重算：只重算还热的评论（heat ≥ 1，老热评时间衰减）；凉透的不再重算（C6）。
// 返回 ticker，app 优雅关闭时调用 Stop。
func StartCommentHeatJob(db *gorm.DB) *time.Ticker {
	recomputeCommentHeat(db)
	ticker := time.NewTicker(HeatInterval)
	go func() {
		for range ticker.C {
			recomputeCommentHeat(db)
		}
	}()
	logger.Info("评论热度/评论数校准任务已启动", zap.Duration("interval", HeatInterval))
	return ticker
}

// recomputeCommentHeat 执行一次校准与重算。
// 软删除的评论（delete_time 非空）不参与计数与热度。
func recomputeCommentHeat(db *gorm.DB) {
	// ① comment_count 校准（软删过滤）
	if err := db.Exec(
		"UPDATE note SET comment_count = (SELECT COUNT(*) FROM comment WHERE target_type = 1 AND target_id = note.id AND delete_time IS NULL)",
	).Error; err != nil {
		logger.Error("评论数校准失败(note)", zap.Error(err))
		return
	}
	if err := db.Exec(
		"UPDATE strategy SET comment_count = (SELECT COUNT(*) FROM comment WHERE target_type = 2 AND target_id = strategy.id AND delete_time IS NULL)",
	).Error; err != nil {
		logger.Error("评论数校准失败(strategy)", zap.Error(err))
		return
	}

	// ② 评论热度重算：只重算"还热的"评论（heat ≥ 1）——
	// 写路径（点赞/发评论/删评论/后台删）已实时重算，本任务唯一职责是老热评的时间衰减；
	// 凉透的（heat < 1）不影响热评榜/置顶，不再重算（方案 C，C6）。
	if err := db.Exec("UPDATE comment SET heat = " + repoComment.RecomputeHeatExpr + " WHERE delete_time IS NULL AND heat >= 1").Error; err != nil {
		logger.Error("评论热度重算失败", zap.Error(err))
		return
	}

	logger.Info("评论热度/评论数校准完成")
}
