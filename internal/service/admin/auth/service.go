package auth

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/auth"
	response "tuyou/internal/model/dto/response/admin/auth"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	"tuyou/internal/repository/admin/auth"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"

	"tuyou/internal/service/admin/auth/b64c"
	"tuyou/pkg/config"
)

type authService struct {
	adminRepo auth.AdminRepository
	redisRepo repository.RedisRepository
}

func NewAdminAuthService(adminRepo auth.AdminRepository, redisRepo repository.RedisRepository) AdminAuthService {
	return &authService{adminRepo: adminRepo, redisRepo: redisRepo}
}

// GetCaptcha 获取图形验证码
func (s *authService) GetCaptcha(ctx context.Context) (*response.CaptchaResp, error) {
	captchaID, b64s, err := b64c.Generate()
	if err != nil {
		return nil, err
	}
	return &response.CaptchaResp{
		CaptchaID:    captchaID,
		CaptchaImage: b64s,
	}, nil
}

// Login 登录，只返回 token，管理员信息通过 GetProfile 获取
func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResp, error) {

	if req.Password == "" {
		return nil, appErrors.ErrMissingParam
	}

	//检验图形验证码
	if err := b64c.Identify(req.CaptchaId, req.Captcha); err != nil {
		return nil, appErrors.ErrCaptchaInvalid
	}

	//检查用户是否存在
	admin, err := s.adminRepo.FindByUsername(ctx, req.UserName)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, err
	}

	//查看用户是否被禁用
	if admin.Status == 0 {
		return nil, appErrors.ErrUserDisabled
	}

	//验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	//异步更新登录时间
	go func() {
		_ = s.adminRepo.UpdateLastLoginTime(context.Background(), admin.ID)
	}()

	//获取登录token
	token, err := jwt.GenerateToken(admin.ID, admin.Username)
	if err != nil {
		return nil, err
	}

	//写入登录态白名单（多端登录：每个 token 独立 key；Redis 不可用时忽略，不影响登录）
	_ = s.redisRepo.Set(ctx, jwt.AdminTokenKey(admin.ID, token), 1, config.Get().JWT.ExpireHours)

	//查询角色名称
	role, err := s.adminRepo.FindRoleNameByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}

	//查询角色菜单与权限
	menus, err := s.adminRepo.FindMenusByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}
	perms, err := s.adminRepo.FindPermsByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}

	return &response.LoginResp{
		Token:    token,
		ID:       strconv.Itoa(int(admin.ID)),
		Username: admin.Username,
		Role:     role,
		Avatar:   admin.Avatar,
		MenuTree: buildMenuTree(menus),
		PermTree: buildPermTree(perms),
	}, nil
}

// buildMenuTree 将扁平菜单列表组装成树
func buildMenuTree(menus []entity.SysMenu) []*response.MenuTree {
	nodes := make(map[uint]*response.MenuTree, len(menus))
	roots := make([]*response.MenuTree, 0, len(menus))
	for i := range menus {
		m := &menus[i]
		nodes[m.ID] = &response.MenuTree{
			ID:        strconv.Itoa(int(m.ID)),
			MenuName:  m.MenuName,
			ParentID:  strconv.Itoa(int(m.ParentID)),
			MenuType:  m.MenuType,
			RoutePath: m.RoutePath,
			Sort:      m.Sort,
		}
	}
	for _, m := range menus {
		node := nodes[m.ID]
		if m.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodes[m.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 父菜单不存在时挂到根，避免节点丢失
			roots = append(roots, node)
		}
	}
	return roots
}

// buildPermTree 将扁平权限列表按模块分组成树
func buildPermTree(perms []entity.SysPermission) []response.PermTree {
	tree := make([]response.PermTree, 0, len(perms))
	idx := make(map[string]int, len(perms))
	for _, p := range perms {
		item := response.PermItem{
			ID:            strconv.Itoa(int(p.ID)),
			PermName:      p.PermName,
			PermKey:       p.PermKey,
			RequestMethod: p.RequestMethod,
			APIPath:       p.APIPath,
		}
		if i, ok := idx[p.Module]; ok {
			tree[i].Children = append(tree[i].Children, item)
		} else {
			idx[p.Module] = len(tree)
			tree = append(tree, response.PermTree{
				Module:   p.Module,
				Children: []response.PermItem{item},
			})
		}
	}
	return tree
}

// GetProfile 获取当前登录管理员信息
func (s *authService) GetProfile(ctx context.Context, adminID uint) (*response.AdminInfo, error) {
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
		Status:    admin.Status == 1,
	}, nil
}

// UpdatePassword 修改当前登录管理员密码
func (s *authService) UpdatePassword(ctx context.Context, adminID uint, req *request.ResetPasswordRequest) error {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		return appErrors.ErrInvalidCredentials
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.adminRepo.UpdatePassword(ctx, adminID, string(hashed))
}
