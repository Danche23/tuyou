package admins

import (
	"context"
	"tuyou/internal/model/dto/response/admin/permission/admins"
	"tuyou/internal/model/entity"
)

type AdminRepository interface {
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	List(ctx context.Context, filters map[string]interface{}, page, size int) ([]admins.AdminListItem, error)
	FindByUsername(ctx context.Context, username string) (*entity.SysAdmin, error)
	FindRoleIDByRoleName(ctx context.Context, roleName string) (uint, error)
	ListRoles(ctx context.Context) ([]admins.RoleItem, error)
	Insert(ctx context.Context, admin *entity.SysAdmin) error
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, password string) error
}
