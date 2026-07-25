package entity

// StrategyLike 攻略点赞记录表
type StrategyLike struct {
	BaseEntity
	StrategyID uint `gorm:"column:strategy_id;not null;uniqueIndex:idx_strategy_member;comment:关联攻略id" json:"strategyId"`
	MemberID   uint `gorm:"column:member_id;not null;uniqueIndex:idx_strategy_member;comment:点赞用户member id" json:"memberId"`
}

// TableName 指定表名
func (StrategyLike) TableName() string {
	return "strategy_like"
}
