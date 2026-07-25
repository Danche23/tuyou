package entity

// NoteLike 笔记点赞记录表
type NoteLike struct {
	BaseEntity
	NoteID   uint `gorm:"column:note_id;not null;uniqueIndex:idx_note_member;comment:关联笔记id" json:"noteId"`
	MemberID uint `gorm:"column:member_id;not null;uniqueIndex:idx_note_member;comment:点赞用户member id" json:"memberId"`
}

// TableName 指定表名
func (NoteLike) TableName() string {
	return "note_like"
}
