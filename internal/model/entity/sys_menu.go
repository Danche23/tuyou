package entity

// SysMenu 后台菜单表
type SysMenu struct {
	BaseEntity
	MenuName  string `gorm:"type:varchar(50);not null;comment:菜单展示名称" json:"menuName"`
	ParentID  uint   `gorm:"column:parent_id;default:0;index;comment:父菜单ID，0代表一级菜单" json:"parentId"`
	MenuType  int    `gorm:"type:tinyint;index;comment:菜单类型：1目录，2页面菜单" json:"menuType"`
	RoutePath string `gorm:"type:varchar(255);comment:前端页面路由地址" json:"routePath"`
	Sort      int    `gorm:"comment:菜单排序序号" json:"sort"`
	Status    int    `gorm:"type:tinyint;default:1;index;comment:0隐藏菜单，1展示菜单" json:"status"`
}

// TableName 指定表名
func (SysMenu) TableName() string {
	return "sys_menu"
}
