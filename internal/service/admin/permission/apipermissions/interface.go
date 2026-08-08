package apipermissions

import (
	"context"
	request "tuyou/internal/model/dto/request/admin/permission/apipermissions"
	response "tuyou/internal/model/dto/response/admin/permission/apipermissions"
)

// ApiPermService 接口权限管理服务接口
type ApiPermService interface {
	// ListApiPerms 查询接口权限列表
	ListApiPerms(ctx context.Context, req *request.ListApiPermsReq) (*response.ListApiPermResp, error)

	// InsertApiPerm 新增接口权限
	InsertApiPerm(ctx context.Context, req *request.InsertApiPermReq) (*response.InsertApiPermResp, error)

	// UpdateApiPerm 编辑接口权限
	UpdateApiPerm(ctx context.Context, id uint, req *request.UpdateApiPermReq) error

	// DeleteApiPerm 删除接口权限（级联清理角色-权限关联）
	DeleteApiPerm(ctx context.Context, id uint) error
}
