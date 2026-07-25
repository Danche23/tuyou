package entity

// SysRolePerm 角色-接口权限关联中间表（多对多）
type SysRolePerm struct {
	BaseEntity
	RoleID uint `gorm:"column:role_id;not null;uniqueIndex:idx_role_perm;comment:关联角色表id" json:"roleId"`
	PermID uint `gorm:"column:perm_id;not null;uniqueIndex:idx_role_perm;comment:关联权限表id" json:"permId"`
}

// TableName 指定表名
func (SysRolePerm) TableName() string {
	return "sys_role_perm"
}
