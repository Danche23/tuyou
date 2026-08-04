package member

import (
	"context"
	"math"

	request "tuyou/internal/model/dto/request/admin/member"
	response "tuyou/internal/model/dto/response/admin/member"
	"tuyou/internal/repository/admin/member"
	appErrors "tuyou/pkg/errors"
)

type memberService struct {
	memberRepo member.MemberRepository
}

// NewMemberService 创建会员服务实例
func NewMemberService(memberRepo member.MemberRepository) MemberService {
	return &memberService{memberRepo: memberRepo}
}

// ListMembers 查询会员列表
func (s *memberService) ListMembers(ctx context.Context, role int, req *request.ListMemberReq) (*response.ListMemberResp, error) {
	filters := map[string]interface{}{"role": role}
	if req.ID != "" {
		filters["id"] = req.ID
	}
	if req.Nickname != "" {
		filters["nickname"] = req.Nickname
	}
	if req.Phone != "" {
		filters["phone"] = req.Phone
	}
	if req.RegisterStart != "" {
		filters["register_start"] = req.RegisterStart
	}
	if req.RegisterEnd != "" {
		filters["register_end"] = req.RegisterEnd + " 23:59:59"
	}

	total, err := s.memberRepo.Count(ctx, filters)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	list, err := s.memberRepo.List(ctx, filters, req.Page, req.Size)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	return &response.ListMemberResp{
		Total:     total,
		Page:      req.Page,
		Size:      req.Size,
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
		List:      list,
	}, nil
}

// GetMemberDetail 获取会员详情
func (s *memberService) GetMemberDetail(ctx context.Context, id uint) (*response.MemberDetail, error) {
	return s.memberRepo.GetDetailByID(ctx, id)
}

// UpdateMember 编辑会员
func (s *memberService) UpdateMember(ctx context.Context, id uint, req *request.UpdateMemberReq) error {
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}

	if len(updates) == 0 {
		return appErrors.ErrMissingParam
	}

	return s.memberRepo.Update(ctx, id, updates)
}

// DeleteMember 删除会员
func (s *memberService) DeleteMember(ctx context.Context, id uint) error {
	return s.memberRepo.Delete(ctx, id)
}
