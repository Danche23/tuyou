package entity

import "time"

// Activity 活动主表（仅领队可发布）
type Activity struct {
	BaseEntity
	LeaderMemberID  uint      `gorm:"column:leader_member_id;not null;index;comment:发布人member表id，代码校验role=0（领队）" json:"leaderMemberId"`
	Title           string    `gorm:"type:varchar(50);not null;comment:活动标题" json:"title"`
	CoverImg        string    `gorm:"type:varchar(255);column:cover_img;comment:活动封面图片地址" json:"coverImg"`
	ActivityContent string    `gorm:"type:text;column:activity_content;comment:活动详情正文" json:"activityContent"`
	Address         string    `gorm:"type:varchar(200);comment:活动线下举办地点" json:"address"`
	IsSensitive     int       `gorm:"type:tinyint;default:0;index;column:is_sensitive;comment:发布时命中敏感词标记(内容仍存原文) 0=否 1=是" json:"isSensitive"`
	StartTime       time.Time `gorm:"column:start_time;not null;index;comment:活动开始时间" json:"startTime"`
	EndTime         time.Time `gorm:"column:end_time;not null;comment:活动结束时间" json:"endTime"`
	SignUpStart     time.Time `gorm:"column:sign_up_start;not null;comment:报名通道开启时间" json:"signUpStart"`
	SignUpEnd       time.Time `gorm:"column:sign_up_end;not null;comment:报名截止时间" json:"signUpEnd"`
	MaxPeople       int       `gorm:"column:max_people;default:20;comment:活动最大容纳报名人数" json:"maxPeople"`
	CurrentPeople   int       `gorm:"column:current_people;default:0;comment:当前已报名人数" json:"currentPeople"`
	Price           float64   `gorm:"type:decimal(10,2);default:0;comment:费用" json:"price"`
	ActivityStatus  int       `gorm:"type:tinyint;column:activity_status;default:1;index;comment:活动状态：1报名中，2进行中，3已结束" json:"activityStatus"`
	IsDel           int       `gorm:"type:tinyint;column:is_del;default:0;comment:逻辑删除 0未删除 1已删除" json:"-"`
	Difficulty      int       `gorm:"type:tinyint;comment:难度（下拉选择 入门、中等、中高等、高等）" json:"difficulty"`
	ActivityType    int       `gorm:"type:tinyint;column:activity_type;index;comment:活动类型" json:"activityType"`
	IsSlide         int       `gorm:"type:tinyint;column:is_slide;default:0;index;comment:是否为首页幻灯片推荐：0=否，1=是" json:"isSlide"`
	WechatQR        string    `gorm:"type:varchar(500);column:wechat_qr;comment:微信群二维码图片地址" json:"wechatQR"`
}

// TableName 指定表名
func (Activity) TableName() string {
	return "activity"
}
