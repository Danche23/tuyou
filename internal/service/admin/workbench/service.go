package workbench

import (
	"context"
	"time"

	response "tuyou/internal/model/dto/response/admin/workbench"
	repoWorkbench "tuyou/internal/repository/admin/workbench"
)

type workbenchService struct {
	repo repoWorkbench.WorkbenchRepository
}

func NewWorkbenchService(repo repoWorkbench.WorkbenchRepository) WorkbenchService {
	return &workbenchService{repo: repo}
}

// GetWorkbench 获取工作台统计数据
func (s *workbenchService) GetWorkbench(ctx context.Context) (*response.WorkbenchResp, error) {
	memberTotal, err := s.repo.CountMembers(ctx)
	if err != nil {
		return nil, err
	}
	activityTotal, err := s.repo.CountActivities(ctx)
	if err != nil {
		return nil, err
	}
	noteTotal, err := s.repo.CountNotes(ctx)
	if err != nil {
		return nil, err
	}
	strategyTotal, err := s.repo.CountStrategies(ctx)
	if err != nil {
		return nil, err
	}
	commentTotal, err := s.repo.CountComments(ctx)
	if err != nil {
		return nil, err
	}

	// 近7天注册：今天往前推6天到今天，共7天
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6)
	trend := make([]response.RegisterTrendItem, 0, 7)
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, i)
		cnt, err := s.repo.CountRegistersByDay(ctx, day)
		if err != nil {
			return nil, err
		}
		trend = append(trend, response.RegisterTrendItem{
			Date:  day.Format("2006-01-02"),
			Count: cnt,
		})
	}

	return &response.WorkbenchResp{
		MemberTotal:   memberTotal,
		ActivityTotal: activityTotal,
		PublishTotal:  noteTotal + strategyTotal,
		CommentTotal:  commentTotal,
		RegisterTrend: trend,
	}, nil
}
