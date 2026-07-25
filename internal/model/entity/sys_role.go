package entity

// SysRole 系统角色表
type SysRole struct {
	BaseEntity
	RoleName    string `gorm:"type:varchar(50);not null;comment:角色名称" json:"roleName"`
	RoleCode    string `gorm:"type:varchar(50);not null;uniqueIndex;comment:角色唯一鉴权编码" json:"roleCode"`
	Description string `gorm:"type:varchar(500);comment:角色功能描述" json:"description"`
	Status      int    `gorm:"type:tinyint;default:1;index;comment:角色状态：0禁用，1正常" json:"status"`
}

// TableName 指定表名
func (SysRole) TableName() string {
	return "sys_role"
}
