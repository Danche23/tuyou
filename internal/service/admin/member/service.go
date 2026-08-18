package member

import (
	"context"
	"math"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/member"
	response "tuyou/internal/model/dto/response/admin/member"
	"tuyou/internal/repository"
	"tuyou/internal/repository/admin/member"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"
)

type memberService struct {
	memberRepo member.MemberRepository
	redisRepo  repository.RedisRepository
}

// NewMemberService 创建会员服务实例
func NewMemberService(memberRepo member.MemberRepository, redisRepo repository.RedisRepository) MemberService {
	return &memberService{memberRepo: memberRepo, redisRepo: redisRepo}
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
	if req.Status != "" {
		v, err := strconv.Atoi(req.Status)
		if err != nil {
			return nil, appErrors.ErrInvalidParam
		}
		filters["status"] = v
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
		// 守卫：已注销会员的状态只能走 /restore 或 /auth/cancel，改绑/禁用/启用都不允许
		detail, err := s.memberRepo.GetDetailByID(ctx, id)
		if err != nil {
			return err
		}
		if detail.Status == 2 {
			return appErrors.New(appErrors.CodeConflict, "已注销会员请使用恢复接口")
		}
		if *req.Status {
			updates["status"] = 1
		} else {
			updates["status"] = 0
			//禁用将用户踢下线
			_ = s.redisRepo.Del(ctx, jwt.MemberTokenKey(id))
		}
	}
	if req.Role != nil {
		switch *req.Role {
		case "领队":
			updates["role"] = 0
		case "普通会员":
			updates["role"] = 1
		default:
			return appErrors.ErrInvalidParam
		}
	}

	if len(updates) == 0 {
		return appErrors.ErrMissingParam
	}

	return s.memberRepo.Update(ctx, id, updates)
}

// DeleteMember 删除会员
func (s *memberService) DeleteMember(ctx context.Context, id uint) error {
	_ = s.redisRepo.Del(ctx, jwt.MemberTokenKey(id))
	return s.memberRepo.Delete(ctx, id)
}

// parseIDs 将字符串 ID 列表转换为 uint 列表
func parseIDs(ids []string) ([]uint, error) {
	result := make([]uint, 0, len(ids))
	for _, s := range ids {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, appErrors.ErrInvalidParam
		}
		result = append(result, uint(id))
	}
	return result, nil
}

// BatchDeleteMembers 批量删除
func (s *memberService) BatchDeleteMembers(ctx context.Context, req *request.BatchDeleteReq) error {
	ids, err := parseIDs(req.IDs)
	if err != nil {
		return err
	}
	for _, id := range ids {
		_ = s.redisRepo.Del(ctx, jwt.MemberTokenKey(id))
	}
	return s.memberRepo.BatchDelete(ctx, ids)
}

// BatchUpdateStatus 批量修改状态
func (s *memberService) BatchUpdateStatus(ctx context.Context, req *request.BatchUpdateStatusReq) error {
	ids, err := parseIDs(req.IDs)
	if err != nil {
		return err
	}
	status := 0
	if req.Status {
		status = 1
	}
	if status == 0 {
		for _, id := range ids {
			_ = s.redisRepo.Del(ctx, jwt.MemberTokenKey(id))
		}
	}
	return s.memberRepo.BatchUpdateStatus(ctx, ids, status)
}

// ExportMembers 导出某角色全部记录
func (s *memberService) ExportMembers(ctx context.Context, role int) ([]response.MemberListItem, error) {
	list, err := s.memberRepo.ListAllByRole(ctx, role)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}
	return list, nil
}
