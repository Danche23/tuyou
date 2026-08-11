package activity

import (
	response "tuyou/internal/model/dto/response/admin/activity"
	"tuyou/pkg/excel"
)

// buildActivityExcel 将活动列表转换为 xlsx 二进制流
func buildActivityExcel(list []response.ActivityListItem) ([]byte, error) {
	headers := []string{"ID", "封面", "标题", "领队ID", "领队名称", "活动类型", "开始时间", "结束时间", "费用", "当前报名", "最大人数", "难度", "是否轮播", "状态", "上下架", "删除时间"}
	rows := make([][]interface{}, 0, len(list))
	for _, item := range list {
		rows = append(rows, []interface{}{
			item.ID, item.CoverImg, item.Title, item.LeaderMemberID, item.LeaderName,
			item.ActivityType, item.StartTime, item.EndTime, item.Price,
			item.CurrentPeople, item.MaxPeople, item.Difficulty,
			item.IsSlide, item.ActivityStatus, item.Status, item.DeleteTime,
		})
	}
	return excel.Build("活动列表", headers, rows)
}
