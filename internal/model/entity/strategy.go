package entity

import "time"

// Strategy 旅行攻略表
type Strategy struct {
	BaseEntity
	MemberID     uint      `gorm:"column:member_id;not null;index;comment:发布用户member表id" json:"memberId"`
	// title+content 共用全文索引 ft_strategy_title_content（ngram 分词），
	// 与 repository/web/search 的 MATCH(title, content) 一一对应，改列需同步。
	Title        string    `gorm:"type:varchar(100);not null;index:ft_strategy_title_content,class:FULLTEXT,option:WITH PARSER ngram,priority:1;comment:攻略标题" json:"title"`
	VideoURL     string    `gorm:"type:varchar(255);column:video_url;comment:视频地址" json:"videoUrl"`
	Content      string    `gorm:"type:text;index:ft_strategy_title_content,priority:2;comment:攻略详情内容" json:"content"`
	ViewCount    int64     `gorm:"type:bigint;default:0;column:view_count;comment:浏览总次数" json:"viewCount"`
	LikeCount    int64     `gorm:"type:bigint;default:0;column:like_count;comment:点赞总数" json:"likeCount"`
	CollectCount int64     `gorm:"type:bigint;default:0;column:collect_count;comment:收藏总数" json:"collectCount"`
	CommentCount int64     `gorm:"type:bigint;default:0;column:comment_count;comment:评论总次数" json:"commentCount"`
	Status       int       `gorm:"type:tinyint;default:1;index;comment:上下架：0下架，1正常" json:"status"`
	PublishTime  time.Time `gorm:"column:publish_time;not null;index;comment:发布时间" json:"publishTime"`
	IsSlide      int       `gorm:"type:tinyint;column:is_slide;default:0;index;comment:是否为首页幻灯片推荐：0=否，1=是" json:"isSlide"`
	Category     int       `gorm:"type:tinyint;index;comment:分类" json:"category"`
	Heat         int       `gorm:"default:0;index;comment:热度（定时更新）" json:"heat"`
	Distance     string    `gorm:"type:varchar(20);comment:距离（单位：公里）" json:"distance"`
	Duration     string    `gorm:"type:varchar(20);comment:总时长" json:"duration"`
	Difficulty   int       `gorm:"type:tinyint;comment:难度（下拉选择 入门、中等、中高等、高等）" json:"difficulty"`
	Transport    string    `gorm:"type:varchar(20);comment:交通方式（用户自由填写）" json:"transport"`
	IsSensitive  int       `gorm:"type:tinyint;default:0;index;column:is_sensitive;comment:发布时命中敏感词标记(内容仍存原文) 0=否 1=是" json:"isSensitive"`
}

// TableName 指定表名
func (Strategy) TableName() string {
	return "strategy"
}
