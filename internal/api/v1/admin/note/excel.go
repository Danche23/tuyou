package note

import (
	"strings"

	response "tuyou/internal/model/dto/response/admin/note"
	"tuyou/pkg/excel"
)

// buildNoteExcel 将笔记列表转换为 xlsx 二进制流
func buildNoteExcel(list []response.NoteListItem) ([]byte, error) {
	headers := []string{"ID", "封面", "标题", "作者", "分类", "标签", "发布时间", "浏览量", "点赞量", "收藏量", "评论量", "热度", "是否轮播", "状态"}
	rows := make([][]interface{}, 0, len(list))
	for _, item := range list {
		rows = append(rows, []interface{}{
			item.ID, item.CoverImg, item.Title, item.Nickname, item.Category,
			strings.Join(item.Tags, ","), item.PublishTime,
			item.ViewCount, item.LikeCount, item.CollectCount, item.CommentCount,
			item.Heat, item.IsSlide, item.Status,
		})
	}
	return excel.Build("笔记列表", headers, rows)
}
