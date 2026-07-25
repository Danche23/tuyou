package entity

// NoteCollect 笔记收藏记录表
type NoteCollect struct {
	BaseEntity
	NoteID   uint `gorm:"column:note_id;not null;uniqueIndex:idx_note_member;comment:关联笔记id" json:"noteId"`
	MemberID uint `gorm:"column:member_id;not null;uniqueIndex:idx_note_member;comment:收藏用户member id" json:"memberId"`
}

// TableName 指定表名
func (NoteCollect) TableName() string {
	return "note_collect"
}
