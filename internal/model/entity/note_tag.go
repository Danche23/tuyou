package entity

// NoteTag 笔记-标签关联表
type NoteTag struct {
	BaseEntity
	NoteID uint `gorm:"column:note_id;not null;uniqueIndex:idx_note_tag;comment:关联笔记表id" json:"noteId"`
	TagID  uint `gorm:"column:tag_id;not null;uniqueIndex:idx_note_tag;comment:关联标签表id" json:"tagId"`
}

// TableName 指定表名
func (NoteTag) TableName() string {
	return "note_tag"
}
