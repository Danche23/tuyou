package entity

// StrategyCollect 攻略收藏记录表
type StrategyCollect struct {
	BaseEntity
	StrategyID uint `gorm:"column:strategy_id;not null;uniqueIndex:idx_strategy_member;comment:关联攻略id" json:"strategyId"`
	MemberID   uint `gorm:"column:member_id;not null;uniqueIndex:idx_strategy_member;comment:收藏用户member id" json:"memberId"`
}

// TableName 指定表名
func (StrategyCollect) TableName() string {
	return "strategy_collect"
}
