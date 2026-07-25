package entity

// StrategyTag 攻略-标签关联表
type StrategyTag struct {
	BaseEntity
	StrategyID uint `gorm:"column:strategy_id;not null;uniqueIndex:idx_strategy_tag;comment:关联攻略表id" json:"strategyId"`
	TagID      uint `gorm:"column:tag_id;not null;uniqueIndex:idx_strategy_tag;comment:关联标签表id" json:"tagId"`
}

// TableName 指定表名
func (StrategyTag) TableName() string {
	return "strategy_tag"
}
