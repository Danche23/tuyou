package center

import (
	"context"

	"gorm.io/gorm"
	"tuyou/internal/model/entity"
)

type centerRepository struct {
	db *gorm.DB
}

func NewCenterRepository(db *gorm.DB) CenterRepository {
	return &centerRepository{db: db}
}

// UpdateAvatar 更新头像
func (r *centerRepository) UpdateAvatar(ctx context.Context, id uint, avatar string) error {
	return r.db.WithContext(ctx).
		Model(&entity.SysAdmin{}).
		Where("id = ?", id).
		Update("avatar", avatar).Error
}

// UpdateUsername 更新用户名
func (r *centerRepository) UpdateUsername(ctx context.Context, id uint, username string) error {
	return r.db.WithContext(ctx).
		Model(&entity.SysAdmin{}).
		Where("id = ?", id).
		Update("username", username).Error
}
