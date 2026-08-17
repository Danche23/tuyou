package member

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/member"
	response "tuyou/internal/model/dto/response/admin/member"
)

// MemberService 会员管理服务接口
type MemberService interface {
	// ListMembers 查询会员列表（role 由路由传入；nil 表示不按角色过滤，会员和领队都展示）
	ListMembers(ctx context.Context, role *int, req *request.ListMemberReq) (*response.ListMemberResp, error)

	// GetMemberDetail 获取会员详情
	GetMemberDetail(ctx context.Context, id uint) (*response.MemberDetail, error)

	// UpdateMember 编辑会员
	UpdateMember(ctx context.Context, id uint, req *request.UpdateMemberReq) error

	// DeleteMember 删除会员
	DeleteMember(ctx context.Context, id uint) error

	// BatchDeleteMembers 批量删除
	BatchDeleteMembers(ctx context.Context, req *request.BatchDeleteReq) error

	// BatchUpdateStatus 批量修改状态
	BatchUpdateStatus(ctx context.Context, req *request.BatchUpdateStatusReq) error

	// ExportMembers 导出某角色全部记录（role 由路由传入）
	ExportMembers(ctx context.Context, role int) ([]response.MemberListItem, error)
}
