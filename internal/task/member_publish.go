package task

import (
	"time"

	"go.uber.org/zap"

	"gorm.io/gorm"

	"tuyou/pkg/logger"
)

// StartMemberPublishJob 启动 member.publish_count 校准定时任务：
// 每小时执行一次，启动时先跑一次。
// publish_count 是"发布数（笔记+攻略）"的冗余计数器，靠笔记/攻略模块的 +1/-1 维护，
// 仍存在漂移可能（种子 SQL 强写、手动改库、维护方新增 bug 等），
// 每小时用 COUNT 从源头重算兜底，与 comment_heat 的 comment_count 校准同理。
// 返回 ticker，app 优雅关闭时调用 Stop。
func StartMemberPublishJob(db *gorm.DB) *time.Ticker {
	recomputeMemberPublish(db)
	ticker := time.NewTicker(HeatInterval)
	go func() {
		for range ticker.C {
			recomputeMemberPublish(db)
		}
	}()
	logger.Info("发布数校准任务已启动", zap.Duration("interval", HeatInterval))
	return ticker
}

// recomputeMemberPublish 执行一次发布数校准：
// publish_count = COUNT(在架笔记) + COUNT(在架攻略)。
// 在架判定与各模块维护语义一致：status=1（上架）AND delete_time IS NULL（未删/未下架）。
// 活动不计入 publish_count（活动的"带队次数"统计在 member.leading_count），
// 这里严格按字段语义只算笔记+攻略，顺带把历史活动污染的错误值纠正回来。
func recomputeMemberPublish(db *gorm.DB) {
	if err := db.Exec(
		"UPDATE member SET publish_count = " +
			"(SELECT COUNT(*) FROM note WHERE member_id = member.id AND status = 1 AND delete_time IS NULL) + " +
			"(SELECT COUNT(*) FROM strategy WHERE member_id = member.id AND status = 1 AND delete_time IS NULL)",
	).Error; err != nil {
		logger.Error("发布数校准失败", zap.Error(err))
		return
	}
	logger.Info("发布数校准完成")
}
