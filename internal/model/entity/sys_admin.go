package entity

import "time"

// SysAdmin 后台管理员账号表
type SysAdmin struct {
	BaseEntity
	Username      string     `gorm:"type:varchar(20);not null;uniqueIndex;comment:后台登录账号" json:"username"`
	Password      string     `gorm:"type:varchar(255);not null;comment:加密登录密码" json:"-"`
	AdminName     string     `gorm:"type:varchar(20);comment:管理员真实姓名" json:"adminName"`
	RoleID        uint       `gorm:"column:role_id;index;comment:关联sys_role角色表id" json:"roleId"`
	Avatar        string     `gorm:"type:varchar(255);comment:管理员头像（路径）" json:"avatar"`
	Status        int        `gorm:"type:tinyint;default:1;index;comment:账号状态：0禁用，1正常" json:"status"`
	LastLoginTime *time.Time `gorm:"column:last_login_time;comment:最后登录时间" json:"lastLoginTime"`
}

// TableName 指定表名
func (SysAdmin) TableName() string {
	return "sys_admin"
}
