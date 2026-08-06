package admins

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"math"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/permission/admins"
	response "tuyou/internal/model/dto/response/admin/permission/admins"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository/admin/permission/admins"
	appErrors "tuyou/pkg/errors"
)

type adminService struct {
	adminRepo admins.AdminRepository
}

func NewAdminService(adminRepo admins.AdminRepository) AdminService {
	return &adminService{adminRepo: adminRepo}
}

// ListAdmin 查询会员列表
func (s *adminService) ListAdmin(ctx context.Context, req *request.ListAdminsReq) (*response.ListAdminResp, error) {
	filters := map[string]interface{}{}
	if req.Username != "" {
		filters["username"] = req.Username
	}
	if req.Role != "" {
		filters["role"] = req.Role
	}
	if req.Status {
		filters["status"] = true
	}

	total, err := s.adminRepo.Count(ctx, filters)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	list, err := s.adminRepo.List(ctx, filters, int(req.Page), int(req.Size))
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	return &response.ListAdminResp{
		Total:     int(total),
		Page:      int(req.Page),
		Size:      int(req.Size),
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
		List:      list,
	}, nil
}

// InsertAdmin 新增管理员
func (s *adminService) InsertAdmin(ctx context.Context, req *request.InsertAdminReq) (*response.InsertAdminResp, error) {
	// 查重：DB 真错误原样返回，已存在返回"已存在"，不存在则继续
	exist, err := s.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, appErrors.ErrUserAlreadyExists
	}

	//将角色名转为对应id
	roleID, err := s.adminRepo.FindRoleIDByRoleName(ctx, req.Role)
	if err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := &entity.SysAdmin{
		Username:  req.Username,
		Password:  string(hashed),
		AdminName: req.Adminname,
		RoleID:    roleID,
		Status:    1,
	}
	if err := s.adminRepo.Insert(ctx, admin); err != nil {
		return nil, err
	}
	return &response.InsertAdminResp{
		Id:       strconv.Itoa(int(admin.ID)),
		Username: admin.Username,
	}, nil
}

// UpdateAdmin 编辑管理员
func (s *adminService) UpdateAdmin(ctx context.Context, id uint, req *request.UpdateAdminReq) error {
	updates := map[string]interface{}{}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Adminname != "" {
		updates["adminname"] = req.Adminname
	}
	if req.Role != "" {
		roleID, err := s.adminRepo.FindRoleIDByRoleName(ctx, req.Role)
		if err != nil {
			return err
		}
		updates["role_id"] = roleID
	}
	if req.Status {
		updates["status"] = 1
	} else {
		updates["status"] = 0
	}

	if len(updates) == 0 {
		return appErrors.ErrMissingParam
	}

	return s.adminRepo.Update(ctx, id, updates)
}

// DeleteAdmin 删除管理员
func (s *adminService) DeleteAdmin(ctx context.Context, id uint) error {
	return s.adminRepo.Delete(ctx, id)
}

// ResetAdminPassword 重置管理员密码
func (s *adminService) ResetAdminPassword(ctx context.Context, id uint, req *request.ResetAdminPasswordReq) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.adminRepo.UpdatePassword(ctx, id, string(hashed))
}
