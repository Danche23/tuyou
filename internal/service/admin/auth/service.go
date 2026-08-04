package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/auth"
	response "tuyou/internal/model/dto/response/admin/auth"
	"tuyou/internal/repository/admin/auth"
	"tuyou/internal/service/admin/auth/b64c"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"
)

type authService struct {
	adminRepo auth.AdminRepository
}

func NewAdminAuthService(adminRepo auth.AdminRepository) AdminAuthService {
	return &authService{adminRepo: adminRepo}
}

// GetCaptcha 获取图形验证码
func (s *authService) GetCaptcha(ctx context.Context) (*response.CaptchaResponse, error) {
	captchaID, b64s, err := b64c.Generate()
	if err != nil {
		return nil, err
	}
	return &response.CaptchaResponse{
		CaptchaID:    captchaID,
		CaptchaImage: b64s,
	}, nil
}

// Login 登录
func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error) {

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

	hashassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(hashassword))
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New(err.Error() + string(hashassword))
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

	//通过角色ID查角色名称
	role, err := s.adminRepo.FindRoleNameByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		Token: token,
		Admin: response.AdminInfo{
			ID:       strconv.Itoa(int(admin.ID)),
			Username: admin.Username,
			Avatar:   admin.Avatar,
			Role:     role,
		},
	}, nil
}
