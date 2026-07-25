package entity

// SysPermission 接口权限表
type SysPermission struct {
	BaseEntity
	PermName      string `gorm:"type:varchar(30);not null;comment:权限名称" json:"permName"`
	Module        string `gorm:"type:varchar(40);index;comment:所属业务模块名称" json:"module"`
	PermKey       string `gorm:"type:varchar(100);uniqueIndex;comment:后端鉴权唯一标识" json:"permKey"`
	RequestMethod string `gorm:"type:varchar(10);comment:请求方式：GET/POST/PUT/DELETE" json:"requestMethod"`
	APIPath       string `gorm:"type:varchar(255);column:api_path;comment:后端接口请求地址" json:"apiPath"`
	Status        int    `gorm:"type:tinyint;default:1;index;comment:0禁用权限，1启用权限" json:"status"`
}

// TableName 指定表名
func (SysPermission) TableName() string {
	return "sys_permission"
}
