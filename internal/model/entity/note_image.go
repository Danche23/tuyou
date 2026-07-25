package entity

// NoteImage 笔记图片表
type NoteImage struct {
	BaseEntity
	NoteID uint   `gorm:"column:note_id;not null;index;comment:关联笔记表id" json:"noteId"`
	URL    string `gorm:"type:varchar(255);not null;comment:图片CDN地址" json:"url"`
	Sort   int    `gorm:"default:0;comment:排序序号，0=封面图（第一张），1=第二张，依次递增" json:"sort"`
}

// TableName 指定表名
func (NoteImage) TableName() string {
	return "note_image"
}
