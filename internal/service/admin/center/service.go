package center

import (
	"context"
	"strconv"

	response "tuyou/internal/model/dto/response/admin/auth"
	repoAuth "tuyou/internal/repository/admin/auth"
	repoCenter "tuyou/internal/repository/admin/center"
)

type centerService struct {
	centerRepo repoCenter.CenterRepository
	adminRepo  repoAuth.AdminRepository
}

func NewCenterService(centerRepo repoCenter.CenterRepository, adminRepo repoAuth.AdminRepository) CenterService {
	return &centerService{centerRepo: centerRepo, adminRepo: adminRepo}
}

// UpdateAvatar 更新头像
func (s *centerService) UpdateAvatar(ctx context.Context, adminID uint, avatar string) (*response.AdminInfo, error) {
	if err := s.centerRepo.UpdateAvatar(ctx, adminID, avatar); err != nil {
		return nil, err
	}
	return s.getProfile(ctx, adminID)
}

// UpdateUsername 更新账号名
func (s *centerService) UpdateUsername(ctx context.Context, adminID uint, username string) (*response.AdminInfo, error) {
	if err := s.centerRepo.UpdateUsername(ctx, adminID, username); err != nil {
		return nil, err
	}
	return s.getProfile(ctx, adminID)
}

// statusStr 将账号状态映射为展示文案：1启用，0禁用
func statusStr(status int) string {
	if status == 1 {
		return "启用"
	}
	return "禁用"
}

// getProfile 查询更新后的管理员信息
func (s *centerService) getProfile(ctx context.Context, adminID uint) (*response.AdminInfo, error) {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	role, err := s.adminRepo.FindRoleNameByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}
	return &response.AdminInfo{
		ID:        strconv.Itoa(int(admin.ID)),
		Username:  admin.Username,
		Adminname: admin.AdminName,
		Role:      role,
		Avatar:    admin.Avatar,
		Status:    statusStr(admin.Status),
	}, nil
}
