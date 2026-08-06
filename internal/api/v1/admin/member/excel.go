package member

import (
	"strconv"

	response "tuyou/internal/model/dto/response/admin/member"

	"github.com/xuri/excelize/v2"
)

// buildExcel 将会员列表转换为 xlsx 二进制流
func buildExcel(list []response.MemberListItem) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "会员列表"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "头像", "昵称", "手机号", "角色", "状态", "简介", "注册时间", "最后登录"}
	for i, h := range headers {
		cell := columnName(i+1) + "1"
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, err
		}
	}

	for idx, item := range list {
		rowNum := idx + 2
		values := []interface{}{
			item.ID, item.Avatar, item.Nickname, item.Phone, item.Role,
			statusText(item.Status), item.Bio, item.RegisterTime, item.LastLogin,
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

// statusText 状态转中文
func statusText(status bool) string {
	if status {
		return "正常"
	}
	return "禁用"
}
