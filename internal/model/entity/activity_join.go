package entity

// ActivityJoin 活动报名记录表
type ActivityJoin struct {
	BaseEntity
	ActivityID uint `gorm:"column:activity_id;not null;uniqueIndex:idx_activity_member;comment:关联活动表id" json:"activityId"`
	MemberID   uint `gorm:"column:member_id;not null;uniqueIndex:idx_activity_member;comment:报名用户member表id，领队/普通会员均可报名" json:"memberId"`
	IsCancel   int  `gorm:"type:tinyint;column:is_cancel;default:0;index;comment:0正常报名，1用户取消报名" json:"isCancel"`
}

// TableName 指定表名
func (ActivityJoin) TableName() string {
	return "activity_join"
}
