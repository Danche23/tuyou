package member

import (
	response "tuyou/internal/model/dto/response/admin/member"
	"tuyou/pkg/excel"
)

// buildExcel 将会员列表转换为 xlsx 二进制流
func buildExcel(list []response.MemberListItem) ([]byte, error) {
	headers := []string{"ID", "头像", "昵称", "手机号", "角色", "状态", "简介", "注册时间", "最后登录"}
	rows := make([][]interface{}, 0, len(list))
	for _, item := range list {
		rows = append(rows, []interface{}{
			item.ID, item.Avatar, item.Nickname, item.Phone, item.Role,
			statusText(item.Status), item.Bio, item.RegisterTime, item.LastLogin,
		})
	}
	return excel.Build("会员列表", headers, rows)
}

// statusText 状态转中文
func statusText(status bool) string {
	if status {
		return "正常"
	}
	return "禁用"
}
