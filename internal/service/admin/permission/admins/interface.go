package admins

import (
	"context"
	request "tuyou/internal/model/dto/request/admin/permission/admins"
	response "tuyou/internal/model/dto/response/admin/permission/admins"
)

// AdminService 管理员管理服务接口
type AdminService interface {
	// ListAdmin 查询管理员列表
	ListAdmin(ctx context.Context, req *request.ListAdminsReq) (*response.ListAdminResp, error)

	// InsertAdmin 新增管理员
	InsertAdmin(ctx context.Context, req *request.InsertAdminReq) (*response.InsertAdminResp, error)

	// UpdateAdmin 编辑管理员
	UpdateAdmin(ctx context.Context, id uint, req *request.UpdateAdminReq) error

	// DeleteAdmin 删除管理员
	DeleteAdmin(ctx context.Context, id uint) error

	// ResetAdminPassword 重置管理员密码
	ResetAdminPassword(ctx context.Context, id uint, req *request.ResetAdminPasswordReq) error
}
