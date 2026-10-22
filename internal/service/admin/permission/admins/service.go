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
	"tuyou/internal/repository"
	"tuyou/internal/repository/admin/permission/admins"
	"tuyou/pkg/config"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"
)

type adminService struct {
	adminRepo admins.AdminRepository
	redisRepo repository.RedisRepository
}

func NewAdminService(adminRepo admins.AdminRepository, redisRepo repository.RedisRepository) AdminService {
	return &adminService{adminRepo: adminRepo, redisRepo: redisRepo}
}

// statusToInt 状态中文转 int：启用=1，禁用=0
func statusToInt(s string) int {
	if s == "禁用" {
		return 0
	}
	return 1
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
	if req.Status != "" {
		filters["status"] = statusToInt(req.Status)
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
	if req.Status != "" {
		updates["status"] = statusToInt(req.Status)
		if updates["status"] == 0 {
			s.kickAdmin(ctx, id)
		}
	}

	if len(updates) == 0 {
		return appErrors.ErrMissingParam
	}

	return s.adminRepo.Update(ctx, id, updates)
}

// kickAdmin 使管理员全部会话立即失效（单端场景即当前会话）：拉黑访问/刷新令牌并清会话标记
func (s *adminService) kickAdmin(ctx context.Context, id uint) {
	cfg := config.Get().JWT

	accessKey := jwt.AdminSessionKey(id)
	var accessJTI string
	if err := s.redisRepo.Get(ctx, accessKey, &accessJTI); err == nil && accessJTI != "" {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, accessJTI), jwt.ReasonLogout, cfg.BlacklistTTL())
		_ = s.redisRepo.Del(ctx, accessKey)
	}

	refreshKey := jwt.RefreshSessionKey(jwt.ScopeAdmin, id)
	var refreshJTI string
	if err := s.redisRepo.Get(ctx, refreshKey, &refreshJTI); err == nil && refreshJTI != "" {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, refreshJTI), jwt.ReasonLogout, cfg.RefreshTTL())
		_ = s.redisRepo.Del(ctx, refreshKey)
	}
}

// DeleteAdmin 删除管理员
func (s *adminService) DeleteAdmin(ctx context.Context, id uint) error {
	s.kickAdmin(ctx, id)
	return s.adminRepo.Delete(ctx, id)
}

// ListRoles 查询角色列表
func (s *adminService) ListRoles(ctx context.Context) ([]response.RoleItem, error) {
	list, err := s.adminRepo.ListRoles(ctx)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}
	return list, nil
}

// ResetAdminPassword 重置管理员密码
func (s *adminService) ResetAdminPassword(ctx context.Context, id uint, req *request.ResetAdminPasswordReq) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.adminRepo.UpdatePassword(ctx, id, string(hashed))
}
