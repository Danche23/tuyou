package member

import (
	"context"

	"gorm.io/gorm"

	response "tuyou/internal/model/dto/response/admin/member"
	appErrors "tuyou/pkg/errors"
)

type memberRepository struct {
	db *gorm.DB
}

// NewMemberRepository 创建会员仓储实例
func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &memberRepository{db: db}
}

// applyFilters 拼装动态查询条件
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["id"]; ok {
		query = query.Where("id = ?", v)
	}
	if v, ok := filters["nickname"]; ok {
		query = query.Where("nickname LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["phone"]; ok {
		query = query.Where("phone LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["register_start"]; ok {
		query = query.Where("create_time >= ?", v)
	}
	if v, ok := filters["register_end"]; ok {
		query = query.Where("create_time <= ?", v)
	}
	if v, ok := filters["role"]; ok {
		query = query.Where("role = ?", v)
	}
	return query
}

// Count 统计数量
func (r *memberRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("member")
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// List 查询会员列表，直接查进
func (r *memberRepository) List(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.MemberListItem, error) {
	var list []response.MemberListItem

	query := r.db.WithContext(ctx).Table("member").
		Select(`id, avatar, nickname,
			CONCAT(LEFT(phone, 3), '****', RIGHT(phone, 4)) AS phone,
			IF(status = 1, TRUE, FALSE) AS status,
			bio, bg_img,
			CASE role WHEN 0 THEN '领队' WHEN 1 THEN '普通会员' ELSE '未知' END AS role,
			DATE_FORMAT(create_time, '%Y-%m-%d') AS register_time,
			IFNULL(DATE_FORMAT(last_login_time, '%Y-%m-%d %H:%i'), '') AS last_login`)
	query = applyFilters(query, filters)

	offset := (page - 1) * size
	if err := query.Order("id DESC").Offset(offset).Limit(size).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetDetailByID 查询会员详情，直接查进
func (r *memberRepository) GetDetailByID(ctx context.Context, id uint) (*response.MemberDetail, error) {
	var detail response.MemberDetail

	err := r.db.WithContext(ctx).Table("member").
		Select(`id, avatar, bg_img, nickname,
			CONCAT(LEFT(phone, 3), '****', RIGHT(phone, 4)) AS phone,
			bio,
			CASE role WHEN 0 THEN '领队' WHEN 1 THEN '普通会员' ELSE '未知' END AS role,
			IF(status = 1, TRUE, FALSE) AS status,
			DATE_FORMAT(create_time, '%Y-%m-%d') AS register_time,
			IFNULL(DATE_FORMAT(last_login_time, '%Y-%m-%d %H:%i'), '') AS last_login,
			follow_count, favorite_count, publish_count, fans_count`).
		Where("id = ?", id).
		Scan(&detail).Error
	if err != nil {
		return nil, err
	}
	if detail.ID == 0 {
		return nil, appErrors.ErrUserNotFound
	}
	return &detail, nil
}

// Update 更新会员字段
func (r *memberRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Table("member").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}

// Delete 硬删除会员
func (r *memberRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Exec("DELETE FROM member WHERE id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrUserNotFound
	}
	return nil
}

// BatchDelete 批量硬删除
func (r *memberRepository) BatchDelete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM member WHERE id IN ?", ids).Error
}

// BatchUpdateStatus 批量更新状态
func (r *memberRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status int) error {
	return r.db.WithContext(ctx).Table("member").
		Where("id IN ?", ids).
		Updates(map[string]interface{}{"status": status}).Error
}

// ListAllByRole 查询某角色全部记录（无分页），用于导出
func (r *memberRepository) ListAllByRole(ctx context.Context, role int) ([]response.MemberListItem, error) {
	var list []response.MemberListItem
	query := r.db.WithContext(ctx).Table("member").
		Select(`id, avatar, nickname,
			CONCAT(LEFT(phone, 3), '****', RIGHT(phone, 4)) AS phone,
			IF(status = 1, TRUE, FALSE) AS status,
			bio, bg_img,
			CASE role WHEN 0 THEN '领队' WHEN 1 THEN '普通会员' ELSE '未知' END AS role,
			DATE_FORMAT(create_time, '%Y-%m-%d') AS register_time,
			IFNULL(DATE_FORMAT(last_login_time, '%Y-%m-%d %H:%i'), '') AS last_login`).
		Where("role = ?", role)
	if err := query.Order("id DESC").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
