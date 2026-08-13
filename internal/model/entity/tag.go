package entity

// Tag 通用标签表（笔记、攻略共用）
type Tag struct {
	BaseEntity
	TagName string `gorm:"type:varchar(30);not null;uniqueIndex;comment:标签名称" json:"tagName"`
	Status  int    `gorm:"type:tinyint;default:1;index;comment:0禁用标签，1启用标签" json:"status"`
	IsHot   int    `gorm:"type:tinyint;default:0;index;comment:0非热门关键词，1热门搜索关键词" json:"isHot"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tag"
}
