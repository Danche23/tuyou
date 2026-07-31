package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
	request "tuyou/internal/model/dto/request/web/auth"
	response "tuyou/internal/model/dto/response/web/auth"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository/web/auth"
	"tuyou/internal/service/web/auth/sms"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"
)

// authService 认证服务实现
type authService struct {
	memberRepo auth.MemberRepository
}

// NewAuthService 创建认证服务
func NewAuthService(memberRepo auth.MemberRepository) WebAuthService {
	return &authService{memberRepo: memberRepo}
}

// Register 注册
func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*response.RegisterResponse, error) {
	//检验手机号格式
	if err := sms.ValidatorPhone(req.Phone); err != nil {
		return nil, err
	}

	//设置默认昵称
	if req.Nickname == "" {
		req.Nickname = fmt.Sprintf("用户%s", req.Phone[8:])
	}

	//校验验证码
	err := sms.Receive(req.Phone, req.Code, sms.SceneRegister)
	if err != nil {
		return nil, appErrors.ErrSmsCodeInvalid
	}

	//检查手机号是否已经被注册
	exist, err := s.memberRepo.FindByPhone(ctx, req.Phone)
	if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, appErrors.ErrUserAlreadyExists
	}

	// bcrypt 加密密码
	hushedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 构建 Member 实体
	now := time.Now()
	member := &entity.Member{
		Nickname:      req.Nickname,
		Phone:         req.Phone,
		Password:      string(hushedPassword),
		Status:        1,
		LastLoginTime: &now,
	}

	// 入库
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	// 生成 JWT token
	token, err := jwt.GenerateToken(member.ID, member.Nickname)
	if err != nil {
		return nil, err
	}

	//返回用户信息
	return &response.RegisterResponse{
		Token: token,
		Member: response.MemberInfo{
			ID:       strconv.Itoa(int(member.ID)),
			Nickname: member.Nickname,
			Phone:    member.Phone,
			Avatar:   member.Avatar,
		},
	}, nil
}

// SendCode 发送验证码
func (s *authService) SendCode(ctx context.Context, req *request.SendCodeRequest) (*response.MessageResponse, error) {
	//校验手机号格式
	if err := sms.ValidatorPhone(req.Phone); err != nil {
		return nil, err
	}

	//按场景做发送端业务检查
	switch sms.Scene(req.Scene) {
	case sms.SceneRegister:
		//注册：手机号已注册则拒发
		exist, err := s.memberRepo.FindByPhone(ctx, req.Phone)
		if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, err
		}
		if exist != nil {
			return nil, appErrors.ErrPhoneAlreadyReg
		}
	case sms.SceneLogin, sms.SceneResetPwd:
		//登录/重置密码：手机号未注册则拒发
		if _, err := s.memberRepo.FindByPhone(ctx, req.Phone); err != nil {
			if errors.Is(err, appErrors.ErrUserNotFound) {
				return nil, appErrors.ErrPhoneNotRegistered
			}
			return nil, err
		}
	}

	return sms.Send(req.Phone, sms.Scene(req.Scene))
}

// Login 登录
func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error) {
	//检验手机号格式
	if err := sms.ValidatorPhone(req.Phone); err != nil {
		return nil, err
	}

	if req.Type == "code" {
		return s.codeLogin(ctx, req)
	}
	return s.passwordLogin(ctx, req)
}

// passwordLogin 密码登录
func (s *authService) passwordLogin(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error) {
	if req.Password == "" {
		return nil, appErrors.ErrMissingParam
	}
	//检查用户是否存在
	member, err := s.memberRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, appErrors.ErrPhoneNotRegistered
		}
		return nil, err
	}

	//检查状态
	if member.Status == 0 {
		return nil, appErrors.ErrUserDisabled
	}

	//验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	//异步来更新登录时间
	go func() {
		_ = s.memberRepo.UpdateLastLoginTime(context.Background(), member.ID)
	}()

	//生成JWT
	token, err := jwt.GenerateToken(member.ID, member.Nickname)
	if err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		Token: token,
		Member: response.MemberInfo{
			ID:       strconv.Itoa(int(member.ID)),
			Nickname: member.Nickname,
			Phone:    member.Phone,
			Avatar:   member.Avatar,
		},
	}, nil
}

// codeLogin验证码登录
func (s *authService) codeLogin(ctx context.Context, req *request.LoginRequest) (*response.LoginResponse, error) {
	if req.Code == "" {
		return nil, appErrors.ErrMissingParam
	}

	//校验验证码
	err := sms.Receive(req.Phone, req.Code, sms.SceneLogin)
	if err != nil {
		return nil, appErrors.ErrSmsCodeInvalid
	}

	member, err := s.memberRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, appErrors.ErrPhoneNotRegistered
		}
		return nil, err
	}

	if member.Status == 0 {
		return nil, appErrors.ErrUserDisabled
	}
	go func() {
		_ = s.memberRepo.UpdateLastLoginTime(context.Background(), member.ID)
	}()

	token, err := jwt.GenerateToken(member.ID, member.Nickname)
	if err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		Token: token,
		Member: response.MemberInfo{
			ID:       strconv.Itoa(int(member.ID)),
			Nickname: member.Nickname,
			Phone:    member.Phone,
			Avatar:   member.Avatar,
		},
	}, nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(ctx context.Context, req *request.ResetPasswordRequest) (*response.MessageResponse, error) {
	//检验手机号格式
	if err := sms.ValidatorPhone(req.Phone); err != nil {
		return nil, err
	}

	//校验验证码
	err := sms.Receive(req.Phone, req.Code, sms.SceneResetPwd)
	if err != nil {
		return nil, appErrors.ErrSmsCodeInvalid
	}

	//查找用户
	member, err := s.memberRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}

	//对新密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := s.memberRepo.UpdatePassword(ctx, member.ID, string(hashedPassword)); err != nil {
		return nil, err
	}
	return &response.MessageResponse{
		Message: "密码重置成功",
	}, nil
}
