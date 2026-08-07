package task

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"tuyou/internal/model/entity"
	"tuyou/pkg/logger"
)

const (
	// HeatInterval 热度重算间隔
	HeatInterval = 1 * time.Hour
	// heatExpr 热度计算公式：
	//   热度 = (浏览量*0.05 + 评论*6 + 点赞*1 + 收藏*3) * 时间系数
	//   时间系数 = 1 / (1 + 发布天数*0.05)，发布越久权重越低，永不归零
	heatExpr = "ROUND((view_count * 0.05 + comment_count * 6 + like_count + collect_count * 3) * (1 / (1 + DATEDIFF(NOW(), publish_time) * 0.05)))"
)

// StartHeatJob 启动热度定时重算任务：启动时立即算一次，之后按间隔循环。
// 返回 ticker，app 优雅关闭时调用 Stop 即可。
func StartHeatJob(db *gorm.DB) *time.Ticker {
	recomputeHeat(db)
	ticker := time.NewTicker(HeatInterval)
	go func() {
		for range ticker.C {
			recomputeHeat(db)
		}
	}()
	logger.Info("热度重算任务已启动", zap.Duration("interval", HeatInterval))
	return ticker
}

// recomputeHeat 按公式批量刷新 note / strategy 两表的 heat 列。
// gorm 默认作用域已过滤软删数据，只更新正常展示的内容。
func recomputeHeat(db *gorm.DB) {
	if err := db.Model(&entity.Note{}).Update("heat", gorm.Expr(heatExpr)).Error; err != nil {
		logger.Error("热度重算失败(note)", zap.Error(err))
		return
	}
	if err := db.Model(&entity.Strategy{}).Update("heat", gorm.Expr(heatExpr)).Error; err != nil {
		logger.Error("热度重算失败(strategy)", zap.Error(err))
		return
	}
	logger.Info("热度重算完成")
}
