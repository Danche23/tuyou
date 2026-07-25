package entity

// MsgTemplate 消息模板表
type MsgTemplate struct {
	BaseEntity
	TemplateName    string `gorm:"type:varchar(50);not null;comment:模板名称" json:"templateName"`
	MsgType         int    `gorm:"type:tinyint;index;comment:主要分两种类型 0 互动、1 系统消息" json:"msgType"`
	ContentTemplate string `gorm:"type:text;column:content_template;comment:消息内容模板，支持变量占位符" json:"contentTemplate"`
	FixedTip        string `gorm:"type:varchar(500);column:fixed_tip;comment:活动通知专用固定提醒文案" json:"fixedTip"`
}

// TableName 指定表名
func (MsgTemplate) TableName() string {
	return "msg_template"
}
