package entity

import "time"

// Member 会员主表（普通会员、领队共用基础信息）
type Member struct {
	BaseEntity
	Avatar        string     `gorm:"type:varchar(255);comment:用户头像图片地址" json:"avatar"`
	Nickname      string     `gorm:"type:varchar(20);not null;comment:用户昵称" json:"nickname"`
	Phone         string     `gorm:"type:varchar(20);uniqueIndex;not null;comment:登录手机号，唯一不重复" json:"phone"`
	Password      string     `gorm:"type:varchar(255);not null;comment:加密存储登录密码（bcrypt）" json:"-"`
	Bio           string     `gorm:"type:varchar(100);comment:用户个人简介/签名" json:"bio"`
	BgImg         string     `gorm:"type:varchar(255);column:bg_img;comment:个人主页背景图路径" json:"bgImg"`
	MemberType    int        `gorm:"type:tinyint;default:2;index;comment:用户类型：0=领队，1=普通会员" json:"memberType"`
	Status        int        `gorm:"type:tinyint;default:1;index;comment:账号状态：0=禁用，1=正常" json:"status"`
	LastLoginTime *time.Time `gorm:"column:last_login_time;comment:最近一次登录时间" json:"lastLoginTime"`
	FollowCount   int64      `gorm:"type:bigint;default:0;comment:关注数" json:"followCount"`
	FavoriteCount int64      `gorm:"type:bigint;default:0;comment:收藏数" json:"favoriteCount"`
	PublishCount  int64      `gorm:"type:bigint;default:0;comment:发布数（笔记+攻略）" json:"publishCount"`
	FansCount     int64      `gorm:"type:bigint;default:0;comment:粉丝数" json:"fansCount"`
	LeadingCount  *int64     `gorm:"type:bigint;default:0;comment:带队次数（发布活动数），默认0，普通会员为NULL" json:"leadingCount"`
}

// TableName 指定表名
func (Member) TableName() string {
	return "member"
}
