package entity

// SensitiveWord 敏感词表（词库来源：开源中文敏感词库导入）
type SensitiveWord struct {
	BaseEntity
	Word   string `gorm:"type:varchar(100);not null;uniqueIndex;comment:敏感词内容" json:"word"`
	Status int    `gorm:"type:tinyint;default:1;index;comment:1启用，0停用" json:"status"`
}

// TableName 指定表名
func (SensitiveWord) TableName() string {
	return "sensitive_word"
}
