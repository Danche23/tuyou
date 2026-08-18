package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	request "tuyou/internal/model/dto/request/web/auth"
	response "tuyou/internal/model/dto/response/web/auth"
	memberResp "tuyou/internal/model/dto/response/web/member"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	"tuyou/internal/repository/web/auth"
	"tuyou/internal/service/web/auth/sms"
	"tuyou/pkg/config"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"
	"tuyou/pkg/sensitive"

	"golang.org/x/crypto/bcrypt"
)

// authService 认证服务实现
type authService struct {
	memberRepo auth.MemberRepository
	redisRepo  repository.RedisRepository
	sensitive  *sensitive.SensitiveMatcher
}

// NewAuthService 创建认证服务
func NewAuthService(memberRepo auth.MemberRepository, redisRepo repository.RedisRepository, matcher *sensitive.SensitiveMatcher) WebAuthService {
	return &authService{memberRepo: memberRepo, redisRepo: redisRepo, sensitive: matcher}
}

// Register 注册
func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*response.MessageResponse, error) {
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
		Avatar:        entity.DefaultAvatar, // 默认头像写入库，保证笔记/攻略/评论等读 member 表的模块都有头像
		Bio:           entity.DefaultBio,
		BgImg:         entity.DefaultBgImg,
		Role:          1, // 前台注册的都是普通会员
		Status:        1,
		LastLoginTime: &now,
	}

	// 入库
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	//注册成功不自动登录，前端引导用户去登录
	return &response.MessageResponse{
		Message: "注册成功",
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
	case sms.SceneRegister, sms.SceneChangePhone:
		//注册/改绑手机号：手机号已注册则拒发
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

	//写入登录态白名单（单端登录：按用户维度覆盖，新登录顶掉旧 token；Redis 不可用时忽略，不影响登录）
	_ = s.redisRepo.Set(ctx, jwt.MemberTokenKey(member.ID), token, config.Get().JWT.ExpireHours)

	return &response.LoginResponse{
		Token: token,
		Member: response.MemberInfo{
			ID:       strconv.Itoa(int(member.ID)),
			Nickname: s.mask(member.Nickname),
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

	//写入登录态白名单（单端登录：按用户维度覆盖，新登录顶掉旧 token；Redis 不可用时忽略，不影响登录）
	_ = s.redisRepo.Set(ctx, jwt.MemberTokenKey(member.ID), token, config.Get().JWT.ExpireHours)

	return &response.LoginResponse{
		Token: token,
		Member: response.MemberInfo{
			ID:       strconv.Itoa(int(member.ID)),
			Nickname: s.mask(member.Nickname),
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

// ChangePhone 修改绑定手机号
func (s *authService) ChangePhone(ctx context.Context, userID uint, req *request.ChangePhoneRequest) (*response.MessageResponse, error) {
	//检验手机号格式
	if err := sms.ValidatorPhone(req.NewPhone); err != nil {
		return nil, err
	}

	//新手机号已注册则拒改（含与原号相同，原号查 phone 必然命中自己）
	exist, err := s.memberRepo.FindByPhone(ctx, req.NewPhone)
	if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, appErrors.ErrPhoneAlreadyReg
	}

	//校验验证码
	if err := sms.Receive(req.NewPhone, req.Code, sms.SceneChangePhone); err != nil {
		return nil, appErrors.ErrSmsCodeInvalid
	}

	//更新手机号
	if err := s.memberRepo.UpdatePhone(ctx, userID, req.NewPhone); err != nil {
		return nil, err
	}
	return &response.MessageResponse{
		Message: "手机号修改成功",
	}, nil
}

// UpdatePassword 登录后修改密码
func (s *authService) UpdatePassword(ctx context.Context, userID uint, req *request.UpdatePasswordRequest) (*response.MessageResponse, error) {
	//查找当前用户
	member, err := s.memberRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	//校验旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.OldPassword)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	//对新密码进行加密并更新
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if err := s.memberRepo.UpdatePassword(ctx, member.ID, string(hashedPassword)); err != nil {
		return nil, err
	}
	return &response.MessageResponse{
		Message: "密码修改成功",
	}, nil
}

// Logout 退出登录：删除该用户白名单登录态，使其立即失效（幂等，Redis 不可用时忽略）
func (s *authService) Logout(ctx context.Context, userID uint) error {
	return s.redisRepo.Del(ctx, jwt.MemberTokenKey(userID))
}

// GetBaseInfo 获取会员基础信息（id/昵称/头像/角色）
func (s *authService) GetBaseInfo(ctx context.Context, userID uint) (*memberResp.MemberBaseInfo, error) {
	member, err := s.memberRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &memberResp.MemberBaseInfo{
		ID:       strconv.Itoa(int(member.ID)),
		Nickname: s.mask(member.Nickname),
		Avatar:   member.Avatar,
		Role:     roleName(member.Role),
	}, nil
}

// roleName 角色数字转名称：0=领队，1=会员
func roleName(role int) string {
	switch role {
	case 0:
		return "领队"
	case 1:
		return "会员"
	default:
		return "未知"
	}
}

// mask 前台文本出口统一打星（nil 安全）
func (s *authService) mask(content string) string {
	if s.sensitive == nil {
		return content
	}
	return s.sensitive.Mask(content)
}
