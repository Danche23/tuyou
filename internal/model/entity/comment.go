package entity

// Comment 评论主表（笔记、攻略共用）
type Comment struct {
	BaseEntity
	TargetType int    `gorm:"type:tinyint;column:target_type;not null;index:idx_target;comment:目标类型：1=笔记评论，2=攻略评论" json:"targetType"`
	TargetID   uint   `gorm:"column:target_id;not null;index:idx_target;comment:关联笔记/攻略ID" json:"targetId"`
	MemberID   uint   `gorm:"column:member_id;not null;index;comment:评论者member表ID" json:"memberId"`
	Content    string `gorm:"type:text;not null;comment:评论内容" json:"content"`
	ReplyToID  *uint  `gorm:"column:reply_to_id;index;comment:回复目标评论ID，NULL=一级评论" json:"replyToId"`
	LikeCount  int64  `gorm:"type:bigint;default:0;column:like_count;comment:点赞总数" json:"likeCount"`
	Status     int    `gorm:"type:tinyint;default:1;index;comment:评论状态：0=下架/删除，1=正常" json:"status"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comment"
}
