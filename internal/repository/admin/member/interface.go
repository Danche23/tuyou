package member

import (
	"context"

	response "tuyou/internal/model/dto/response/admin/member"
)

// MemberRepository 会员仓储接口
type MemberRepository interface {
	// Count 统计数量
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)

	// List 查询会员列表，直接返回
	List(ctx context.Context, filters map[string]interface{}, page, size int) ([]response.MemberListItem, error)

	// GetDetailByID 查询会员详情，直接返回
	GetDetailByID(ctx context.Context, id uint) (*response.MemberDetail, error)

	// Update 更新会员字段
	Update(ctx context.Context, id uint, updates map[string]interface{}) error

	// Delete 硬删除会员
	Delete(ctx context.Context, id uint) error
}
