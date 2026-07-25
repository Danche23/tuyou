package entity

// CommentLike 评论点赞记录表
type CommentLike struct {
	BaseEntity
	CommentID uint `gorm:"column:comment_id;not null;uniqueIndex:idx_comment_member;comment:关联评论表ID" json:"commentId"`
	MemberID  uint `gorm:"column:member_id;not null;uniqueIndex:idx_comment_member;comment:点赞用户member表ID" json:"memberId"`
}

// TableName 指定表名
func (CommentLike) TableName() string {
	return "comment_like"
}
