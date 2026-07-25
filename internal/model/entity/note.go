package entity

import "time"

// Note 游记笔记表
type Note struct {
	BaseEntity
	MemberID     uint      `gorm:"column:member_id;not null;index;comment:发布用户member表id，领队/普通会员均可发布" json:"memberId"`
	Title        string    `gorm:"type:varchar(100);not null;comment:笔记标题" json:"title"`
	VideoURL     string    `gorm:"type:varchar(255);column:video_url;comment:视频地址" json:"videoUrl"`
	Content      string    `gorm:"type:text;comment:笔记正文内容" json:"content"`
	ViewCount    int64     `gorm:"type:bigint;default:0;column:view_count;comment:笔记浏览总次数" json:"viewCount"`
	LikeCount    int64     `gorm:"type:bigint;default:0;column:like_count;comment:笔记点赞总数" json:"likeCount"`
	CollectCount int64     `gorm:"type:bigint;default:0;column:collect_count;comment:笔记收藏总数" json:"collectCount"`
	CommentCount int64     `gorm:"type:bigint;default:0;column:comment_count;comment:笔记评论总数" json:"commentCount"`
	Status       int       `gorm:"type:tinyint;default:1;index;comment:上下架：0下架，1正常展示" json:"status"`
	PublishTime  time.Time `gorm:"column:publish_time;not null;index;comment:发布时间" json:"publishTime"`
	IsSlide      int       `gorm:"type:tinyint;column:is_slide;default:0;index;comment:是否为首页幻灯片推荐：0=否，1=是" json:"isSlide"`
	Category     int       `gorm:"type:tinyint;index;comment:分类" json:"category"`
	Heat         int       `gorm:"default:0;index;comment:热度（定时更新）" json:"heat"`
}

// TableName 指定表名
func (Note) TableName() string {
	return "note"
}
