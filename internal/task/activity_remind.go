package task

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"tuyou/internal/model/entity"
	serviceMsg "tuyou/internal/service/web/message"
	"tuyou/pkg/logger"
)

const (
	// RemindInterval 活动快开始提醒检查间隔
	RemindInterval = 10 * time.Minute
)

// StartActivityRemindJob 启动活动快开始提醒任务：启动时立即检查一次，之后按间隔循环。
// 触发时机：报名截止（sign_up_end 已到）且活动未开始 → 给已报名者发一次提醒。
// 返回 ticker，app 优雅关闭时调用 Stop 即可。
func StartActivityRemindJob(db *gorm.DB, notifier *serviceMsg.Notifier) *time.Ticker {
	remindActivities(db, notifier)
	ticker := time.NewTicker(RemindInterval)
	go func() {
		for range ticker.C {
			remindActivities(db, notifier)
		}
	}()
	logger.Info("活动快开始提醒任务已启动", zap.Duration("interval", RemindInterval))
	return ticker
}

// remindActivities 扫描「报名已截止且未开始」的活动，逐个给已报名未取消的用户发提醒。
func remindActivities(db *gorm.DB, notifier *serviceMsg.Notifier) {
	now := time.Now()
	var activities []entity.Activity
	if err := db.Where("is_del = 0 AND start_time > ? AND sign_up_end <= ?", now, now).
		Find(&activities).Error; err != nil {
		logger.Error("活动快开始提醒: 查询活动失败", zap.Error(err))
		return
	}
	for i := range activities {
		remindActivity(db, notifier, &activities[i])
	}
}

// remindActivity 给一个活动的所有已报名用户发"快开始"提醒；该提醒内容已入库过则跳过（幂等）。
func remindActivity(db *gorm.DB, notifier *serviceMsg.Notifier, a *entity.Activity) {
	content := "你报名的活动《" + a.Title + "》将于 " + a.StartTime.Format("2006-01-02 15:04") + " 开始"

	// 幂等：同一条提醒内容已存在 → 已发过，跳过（每个活动只提醒一次）
	var cnt int64
	if err := db.Model(&entity.MsgInfo{}).
		Where("is_del = 0 AND msg_type = ? AND msg_content = ?", 2, content).
		Count(&cnt).Error; err != nil {
		logger.Error("活动快开始提醒: 查重失败", zap.Uint("activityId", a.ID), zap.Error(err))
		return
	}
	if cnt > 0 {
		return
	}

	var joins []entity.ActivityJoin
	if err := db.Where("activity_id = ? AND is_cancel = 0", a.ID).Find(&joins).Error; err != nil {
		logger.Error("活动快开始提醒: 查询报名者失败", zap.Uint("activityId", a.ID), zap.Error(err))
		return
	}
	for _, j := range joins {
		notifier.NotifyActivityStart(j.MemberID, content)
	}
}
