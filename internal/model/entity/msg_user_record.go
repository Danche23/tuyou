package entity

// MsgUserRecord 用户消息阅读接收记录表
type MsgUserRecord struct {
	BaseEntity
	MsgID           uint `gorm:"column:msg_id;not null;index;comment:关联消息主表id" json:"msgId"`
	ReceiveMemberID uint `gorm:"column:receive_member_id;not null;index;comment:接收消息的会员id" json:"receiveMemberId"`
	IsRead          int  `gorm:"type:tinyint;column:is_read;default:0;index;comment:是否已读：0未读，1已读，默认0" json:"isRead"`
	IsUserDel       int  `gorm:"type:tinyint;column:is_user_del;default:0;comment:用户是否删除消息：0正常，1用户删除" json:"-"`
}

// TableName 指定表名
func (MsgUserRecord) TableName() string {
	return "msg_user_record"
}
