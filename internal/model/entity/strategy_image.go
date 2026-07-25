package entity

// StrategyImage 攻略图片表
type StrategyImage struct {
	BaseEntity
	StrategyID uint   `gorm:"column:strategy_id;not null;index;comment:关联攻略表id" json:"strategyId"`
	URL        string `gorm:"type:varchar(255);not null;comment:图片CDN地址" json:"url"`
	Sort       int    `gorm:"default:0;comment:排序序号，0=封面图（第一张），1=第二张，依次递增" json:"sort"`
}

// TableName 指定表名
func (StrategyImage) TableName() string {
	return "strategy_image"
}
