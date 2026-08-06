package note

import (
	"strconv"
	"strings"

	response "tuyou/internal/model/dto/response/admin/note"

	"github.com/xuri/excelize/v2"
)

// buildNoteExcel 将笔记列表转换为 xlsx 二进制流
func buildNoteExcel(list []response.NoteListItem) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "笔记列表"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "封面", "标题", "作者", "分类", "标签", "发布时间", "浏览量", "点赞量", "收藏量", "评论量", "热度", "是否轮播", "状态"}
	for i, h := range headers {
		cell := columnName(i+1) + "1"
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, err
		}
	}

	for idx, item := range list {
		rowNum := idx + 2
		values := []interface{}{
			item.ID, item.CoverImg, item.Title, item.Nickname, item.Category,
			strings.Join(item.Tags, ","), item.PublishTime,
			item.ViewCount, item.LikeCount, item.CollectCount, item.CommentCount,
			item.Heat, item.IsSlide, item.Status,
		}
		for i, v := range values {
			cell := columnName(i+1) + strconv.Itoa(rowNum)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				return nil, err
			}
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// columnName 生成 Excel 列名（A、B、...、Z、AA...）
func columnName(col int) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+col%26)) + name
		col /= 26
	}
	return name
}
