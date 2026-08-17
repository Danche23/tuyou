package message

// ListMessageReq 后台消息分页查询请求（互动/系统共用）
type ListMessageReq struct {
	Page       int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size       int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	Type       string `form:"type"`                                  // 消息类型筛选：互动=点赞/关注/评论/收藏；系统=活动/系统通知
	BeginTime  string `form:"begin_time"`                            // 发送时间起始 yyyy-MM-dd
	FinishTime string `form:"finish_time"`                           // 发送时间结束 yyyy-MM-dd
}

// PublishSystemMsgReq 发布系统消息请求
type PublishSystemMsgReq struct {
	Content     string `json:"content" binding:"required"` // 消息正文 msg_content
	ReceiveType int    `json:"receive_type"`               // 接收类型：0=全体，1=指定会员（对应 receive_obj_type）
	MemberIDs   []uint `json:"member_ids"`                 // receive_type=1 时的接收会员id列表
}
