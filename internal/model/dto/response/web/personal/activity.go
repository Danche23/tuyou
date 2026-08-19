package personal

import (
	responseActivity "tuyou/internal/model/dto/response/web/activity"
)

// ParticipantInfo 活动报名人（手机号明文，仅活动领队本人可见；昵称打星后返回）
type ParticipantInfo struct {
	ID       string `json:"id"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

// PublishActivityItem 我发布的活动列表项 = 活动列表项 + 报名人（全量，前端弹窗内自行分页）
type PublishActivityItem struct {
	responseActivity.ActivityListItem
	Participants []ParticipantInfo `json:"participants"`
}
