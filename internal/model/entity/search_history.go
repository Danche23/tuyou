package entity

// SearchHistory 搜索历史表（按用户独立，去重置顶，每人最多保留最近 20 条）
type SearchHistory struct {
	BaseEntity
	MemberID uint   `gorm:"column:member_id;not null;uniqueIndex:uk_member_keyword;comment:用户member表id" json:"member_id"` // 用户ID
	Keyword  string `gorm:"type:varchar(50);not null;uniqueIndex:uk_member_keyword;comment:搜索关键词" json:"keyword"`         // 搜索关键词
}

// TableName 表名
func (SearchHistory) TableName() string {
	return "search_history"
}
