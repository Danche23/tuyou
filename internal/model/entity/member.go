package entity

import "time"

// 会员资料默认值：注册时写入 member 表。
// 这样所有读 member 表的模块（笔记/攻略/评论/个人主页）都能取到非空头像与背景图，
// 避免只在某一模块做常量兜底导致其他模块（新用户发布笔记/攻略）取到空头像。
const (
	DefaultAvatar = "https://tuyou1.oss-cn-beijing.aliyuncs.com/images/2026/08/10/%E9%BB%98%E8%AE%A4%E5%A4%B4%E5%83%8F.png"                 // 默认头像：小图
	DefaultBio    = "这个人很懒，什么都没留下"                                                                                             // 默认简介
	DefaultBgImg  = "https://tuyou1.oss-cn-beijing.aliyuncs.com/images/2026/08/10/%E9%BB%98%E8%AE%A4%E8%83%8C%E6%99%AF%E5%9B%BE.jpg" // 默认背景图：大图
)

// Member 会员主表（普通会员、领队共用基础信息）
type Member struct {
	BaseEntity
	Avatar        string     `gorm:"type:varchar(255);comment:用户头像图片地址" json:"avatar"`
	Nickname      string     `gorm:"type:varchar(20);not null;comment:用户昵称" json:"nickname"`
	Phone         string     `gorm:"type:varchar(20);not null;comment:登录手机号，注销后可重新注册（唯一性改代码级校验）" json:"phone"`
	Password      string     `gorm:"type:varchar(255);not null;comment:加密存储登录密码（bcrypt）" json:"-"`
	Bio           string     `gorm:"type:varchar(100);comment:用户个人简介/签名" json:"bio"`
	BgImg         string     `gorm:"type:varchar(255);column:bg_img;comment:个人主页背景图路径" json:"bgImg"`
	Role          int        `gorm:"type:tinyint;default:2;index;comment:用户类型：0=领队，1=普通会员" json:"role"`
	Status        int        `gorm:"type:tinyint;default:1;index;comment:账号状态：0=禁用，1=正常，2=已注销" json:"status"`
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
