package entity

import "time"

// MsgInfo 消息主表（系统推送消息）
type MsgInfo struct {
	BaseEntity
	MsgType        int       `gorm:"type:tinyint;index;comment:1=互动消息，2=系统通知" json:"msgType"`
	ReceiveObjType int       `gorm:"type:tinyint;column:receive_obj_type;index;comment:接收对象：0全体系统用户，1指定会员" json:"receiveObjType"`
	SenderID       uint      `gorm:"column:sender_id;default:0;index;comment:发送者ID：0=系统，其他为管理员/会员id" json:"senderId"`
	TemplateID     uint      `gorm:"column:template_id;index;comment:关联消息模板id" json:"templateId"`
	MsgContent     string    `gorm:"type:text;column:msg_content;comment:渲染完成后的最终消息正文" json:"msgContent"`
	SendTime       time.Time `gorm:"column:send_time;index;comment:消息生成发送时间" json:"sendTime"`
	IsDel          int       `gorm:"type:tinyint;column:is_del;default:0;comment:逻辑删除（后台新添加）" json:"-"`
}

// TableName 指定表名
func (MsgInfo) TableName() string {
	return "msg_info"
}
