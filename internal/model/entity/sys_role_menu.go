package entity

// SysRoleMenu 角色-菜单关联中间表（多对多）
type SysRoleMenu struct {
	BaseEntity
	RoleID uint `gorm:"column:role_id;not null;uniqueIndex:idx_role_menu;comment:关联角色表id" json:"roleId"`
	MenuID uint `gorm:"column:menu_id;not null;uniqueIndex:idx_role_menu;comment:关联菜单表id" json:"menuId"`
}

// TableName 指定表名
func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}
