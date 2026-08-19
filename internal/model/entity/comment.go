package entity

// Comment 评论主表（笔记、攻略共用）
type Comment struct {
	BaseEntity
	TargetType      int    `gorm:"type:tinyint;column:target_type;not null;index:idx_target;comment:目标类型：1=笔记评论，2=攻略评论" json:"targetType"`
	TargetID        uint   `gorm:"column:target_id;not null;index:idx_target;comment:关联笔记/攻略ID" json:"targetId"`
	MemberID        uint   `gorm:"column:member_id;not null;index;comment:评论者member表ID" json:"memberId"`
	Content         string `gorm:"type:text;not null;comment:评论内容" json:"content"`
	ReplyToID       *uint  `gorm:"column:reply_to_id;index;comment:回复目标评论ID，NULL=一级评论" json:"replyToId"`
	ReplyToMemberID *uint  `gorm:"column:reply_to_member_id;index;comment:被回复的用户member表ID，NULL=一级评论无被回复用户" json:"replyToMemberId"`
	ParentID        *uint  `gorm:"column:parent_id;index;comment:回复的直接父评论ID（真正回复的那条），NULL=一级评论；消息中心回复列表按此绑定" json:"parentId"`
	LikeCount       int64  `gorm:"type:bigint;default:0;column:like_count;comment:点赞总数" json:"likeCount"`
	Heat            int64  `gorm:"type:bigint;default:0;column:heat;comment:热度值(互动分×时间因子+新鲜度加分)，写操作时异步更新" json:"heat"`
	HeatScore       int64  `gorm:"type:bigint;default:0;column:heat_score;comment:累计热度分(只增不减，写路径事务内按配额/上限累加)" json:"heatScore"`
	HeatScored      int    `gorm:"type:tinyint;default:0;column:heat_scored;comment:二级评论标记：该回复是否已计入父级热度(1=已计入)" json:"heatScored"`
	IsSensitive     int    `gorm:"type:tinyint;default:0;column:is_sensitive;comment:发布时敏感词快照：1=命中敏感词(内容仍存原文)" json:"isSensitive"`
	HitWords        string `gorm:"type:varchar(255);default:'';column:hit_words;comment:命中的敏感词，分号分隔(发布时快照)" json:"hitWords"`
	Status          int    `gorm:"type:tinyint;default:1;index;comment:评论状态：0=下架/删除，1=正常" json:"status"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comment"
}
