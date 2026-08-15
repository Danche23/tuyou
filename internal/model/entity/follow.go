package entity

// Follow 用户关注表
type Follow struct {
	BaseEntity
	FollowerID uint `gorm:"column:follower_id;not null;uniqueIndex:idx_follower_followee;comment:关注者 member.id" json:"followerId"`
	FolloweeID uint `gorm:"column:followee_id;not null;uniqueIndex:idx_follower_followee;index;comment:被关注者 member.id" json:"followeeId"`
	// Status 关注状态位：1=已关注，0=已取关。方案C：关注/取关不再物理删除，用位翻转表达。
	Status int `gorm:"type:tinyint;not null;default:1;comment:关注状态：1=已关注，0=已取关" json:"status"`
}

// TableName 指定表名
func (Follow) TableName() string {
	return "follow"
}
