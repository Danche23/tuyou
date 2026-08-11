package workbench

import (
	"context"
	"time"
)

// WorkbenchRepository 后台工作台数据仓储
type WorkbenchRepository interface {
	// CountMembers 用户总数（不含软删除）
	CountMembers(ctx context.Context) (int64, error)

	// CountActivities 活动总数（不含逻辑删除 is_del=1）
	CountActivities(ctx context.Context) (int64, error)

	// CountNotes 笔记总数（不含软删除）
	CountNotes(ctx context.Context) (int64, error)

	// CountStrategies 攻略总数（不含软删除）
	CountStrategies(ctx context.Context) (int64, error)

	// CountComments 评论总数（仅正常 status=1，不含软删除）
	CountComments(ctx context.Context) (int64, error)

	// CountRegistersByDay 统计某天（[day, day+1)）注册人数
	CountRegistersByDay(ctx context.Context, day time.Time) (int64, error)
}
