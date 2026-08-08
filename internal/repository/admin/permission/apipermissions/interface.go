package apipermissions

import (
	"context"
	response "tuyou/internal/model/dto/response/admin/permission/apipermissions"
	"tuyou/internal/model/entity"
)

type ApiPermRepository interface {
	// Count 计算总数
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	// List 分页查询接口权限列表
	List(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.ApiPermListItem, error)
	// FindByPermKey 按权限标识查询（唯一性校验）
	FindByPermKey(ctx context.Context, permKey string) (*entity.SysPermission, error)
	// FindByID 按 ID 查询接口权限
	FindByID(ctx context.Context, id uint) (*entity.SysPermission, error)
	// Insert 新增接口权限
	Insert(ctx context.Context, perm *entity.SysPermission) error
	// Update 更新接口权限字段
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	// DeleteCascade 事务内删除接口权限及角色-权限关联
	DeleteCascade(ctx context.Context, id uint) error
}
