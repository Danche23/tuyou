package apipermissions

import (
	"context"
	"errors"

	"gorm.io/gorm"
	response "tuyou/internal/model/dto/response/admin/permission/apipermissions"
	"tuyou/internal/model/entity"
	appErrors "tuyou/pkg/errors"
)

type apiPermRepository struct {
	db *gorm.DB
}

func NewApiPermRepository(db *gorm.DB) ApiPermRepository {
	return &apiPermRepository{db: db}
}

// applyFilters 拼装动态查询条件
func applyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	if v, ok := filters["perm_name"]; ok {
		query = query.Where("p.perm_name LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["module"]; ok {
		query = query.Where("p.module LIKE ?", "%"+v.(string)+"%")
	}
	if v, ok := filters["status"]; ok {
		query = query.Where("p.status = ?", v)
	}
	return query
}

// Count 计算总数
func (r *apiPermRepository) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Table("sys_permission AS p")
	query = applyFilters(query, filters)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// List 分页查询接口权限列表
func (r *apiPermRepository) List(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.ApiPermListItem, error) {
	var list []response.ApiPermListItem

	query := r.db.WithContext(ctx).Table("sys_permission AS p").
		Select(`p.id, p.perm_name, p.module, p.perm_key, p.request_method, p.api_path,
			IF(p.status = 1, '启用', '禁用') AS status,
			DATE_FORMAT(p.create_time, '%Y-%m-%d') AS create_time`)
	query = applyFilters(query, filters)

	offset := (page - 1) * size
	if err := query.Order("p.id ASC").Offset(offset).Limit(size).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByPermKey 按权限标识查询（唯一性校验）
func (r *apiPermRepository) FindByPermKey(ctx context.Context, permKey string) (*entity.SysPermission, error) {
	var perm entity.SysPermission
	err := r.db.WithContext(ctx).Where("perm_key = ?", permKey).First(&perm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

// FindByID 按 ID 查询接口权限
func (r *apiPermRepository) FindByID(ctx context.Context, id uint) (*entity.SysPermission, error) {
	var perm entity.SysPermission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&perm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

// Insert 新增接口权限
func (r *apiPermRepository) Insert(ctx context.Context, perm *entity.SysPermission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

// Update 更新接口权限字段
func (r *apiPermRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Table("sys_permission").Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrNotFound
	}
	return nil
}

// DeleteCascade 事务内先删角色-接口权限关联，再删接口权限
func (r *apiPermRepository) DeleteCascade(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_perm WHERE perm_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sys_permission WHERE id = ?", id).Error; err != nil {
			return err
		}
		return nil
	})
}
