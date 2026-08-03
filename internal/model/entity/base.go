package entity

import (
	"time"

	"gorm.io/gorm"
)

// BaseEntity 基础实体
type BaseEntity struct {
	ID        uint           `gorm:"primaryKey" json:"id"`                 //ID
	CreatedAt time.Time      `gorm:"column:create_time" json:"created_at"` //创建时间
	UpdatedAt time.Time      `gorm:"column:update_time" json:"updated_at"` //更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:delete_time;index" json:"-"`    //删除时间
}

// BeforeCreate GORM hook - 创建前
func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate GORM hook - 更新前
func (b *BaseEntity) BeforeUpdate(tx *gorm.DB) error {
	return nil
}
