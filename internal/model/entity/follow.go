package entity

// Follow 用户关注表
type Follow struct {
	BaseEntity
	FollowerID uint `gorm:"column:follower_id;not null;uniqueIndex:idx_follower_followee;comment:关注者 member.id" json:"followerId"`
	FolloweeID uint `gorm:"column:followee_id;not null;uniqueIndex:idx_follower_followee;index;comment:被关注者 member.id" json:"followeeId"`
}

// TableName 指定表名
func (Follow) TableName() string {
	return "follow"
}
