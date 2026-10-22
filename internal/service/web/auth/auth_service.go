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
	"tuyou/internal/service/admin/auth/b64c"
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

	//检查手机号是否已经被注册（仅看在用的 status<>2；已注销号可重新注册）
	exist, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone)
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
		//注册/改绑手机号：在用的手机号已注册则拒发（已注销号不再占用）
		exist, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone)
		if err != nil && !errors.Is(err, appErrors.ErrUserNotFound) {
			return nil, err
		}
		if exist != nil {
			return nil, appErrors.ErrPhoneAlreadyReg
		}
	case sms.SceneLogin, sms.SceneResetPwd:
		//登录/重置密码：手机号未注册则拒发
		if _, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone); err != nil {
			if errors.Is(err, appErrors.ErrUserNotFound) {
				return nil, appErrors.ErrPhoneNotRegistered
			}
			return nil, err
		}
	case sms.SceneCancelAccount:
		//注销：手机号必须已注册（同 login/reset-pwd），且非已注销（已注销拒发）
		if _, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone); err != nil {
			if errors.Is(err, appErrors.ErrUserNotFound) {
				if _, cerr := s.memberRepo.FindCancelledByPhone(ctx, req.Phone); cerr == nil {
					return nil, appErrors.ErrAccountCancelled // 已注销拒发
				}
				return nil, appErrors.ErrPhoneNotRegistered
			}
			return nil, err
		}
	}

	return sms.Send(req.Phone, sms.Scene(req.Scene))
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
	//校验图形验证码
	if err := b64c.Identify(req.CaptchaId, req.Captcha); err != nil {
		return nil, appErrors.ErrCaptchaInvalid
	}
	//检查用户是否存在（仅查"在用的手机号"；若该号只剩已注销账号则区分提示）
	member, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			if _, cerr := s.memberRepo.FindCancelledByPhone(ctx, req.Phone); cerr == nil {
				return nil, appErrors.ErrAccountCancelled // 已注销账号，明确提示
			}
			return nil, appErrors.ErrPhoneNotRegistered
		}
		return nil, err
	}

	//检查状态：禁用拒登；已注销拒登（防御：FindActiveByPhone 已过滤 status=2，这里显式留一道）
	if member.Status == 0 {
		return nil, appErrors.ErrUserDisabled
	}
	if member.Status == 2 {
		return nil, appErrors.ErrAccountCancelled
	}

	//验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	//异步来更新登录时间
	go func() {
		_ = s.memberRepo.UpdateLastLoginTime(context.Background(), member.ID)
	}()

	//签发访问令牌 + 刷新令牌
	pair, err := jwt.GenerateTokenPair(member.ID, member.Nickname, jwt.ScopeMember)
	if err != nil {
		return nil, err
	}

	//单端黑名单：新登录顶号——拉黑上一会话的访问/刷新令牌后覆盖会话标记（Redis 不可用时忽略，不影响登录）
	s.startSession(ctx, member.ID, pair.AccessJTI, pair.RefreshJTI)

	return s.loginResponse(pair, member), nil
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

	member, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			if _, cerr := s.memberRepo.FindCancelledByPhone(ctx, req.Phone); cerr == nil {
				return nil, appErrors.ErrAccountCancelled
			}
			return nil, appErrors.ErrPhoneNotRegistered
		}
		return nil, err
	}

	if member.Status == 0 {
		return nil, appErrors.ErrUserDisabled
	}
	if member.Status == 2 {
		return nil, appErrors.ErrAccountCancelled // 防御
	}
	go func() {
		_ = s.memberRepo.UpdateLastLoginTime(context.Background(), member.ID)
	}()

	//签发访问令牌 + 刷新令牌
	pair, err := jwt.GenerateTokenPair(member.ID, member.Nickname, jwt.ScopeMember)
	if err != nil {
		return nil, err
	}

	//单端黑名单：新登录顶号——拉黑上一会话的访问/刷新令牌后覆盖会话标记（Redis 不可用时忽略，不影响登录）
	s.startSession(ctx, member.ID, pair.AccessJTI, pair.RefreshJTI)

	return s.loginResponse(pair, member), nil
}

// loginResponse 组装登录响应（Token 字段兼容旧客户端，等同于 AccessToken）
func (s *authService) loginResponse(pair *jwt.TokenPair, member *entity.Member) *response.LoginResponse {
	return &response.LoginResponse{
		Token:            pair.AccessToken,
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        pair.AccessExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
		Member: response.MemberInfo{
			ID:       strconv.Itoa(int(member.ID)),
			Nickname: s.mask(member.Nickname),
			Phone:    member.Phone,
			Avatar:   member.Avatar,
		},
	}
}

// Refresh 使用刷新令牌换取新的令牌对（刷新令牌一次性使用，旧令牌立即拉黑轮换）
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*response.RefreshResponse, error) {
	claims, err := jwt.ParseRefreshTokenForScope(refreshToken, jwt.ScopeMember)
	if err != nil {
		return nil, tokenError(err)
	}

	if s.isBlacklisted(ctx, jwt.ScopeMember, claims.ID) {
		return nil, appErrors.ErrInvalidToken
	}

	//刷新前确认账号仍然有效，避免已禁用/注销账号继续续期
	member, err := s.memberRepo.FindByID(ctx, claims.UserID)
	if err != nil || member.Status != 1 {
		return nil, appErrors.ErrInvalidToken
	}

	//校验当前刷新令牌：Redis 可用时，会话中记录的 jti 必须与本次令牌一致，防止旧令牌重放
	if s.redisRepo.Available(ctx) {
		var current string
		if err := s.redisRepo.Get(ctx, jwt.RefreshSessionKey(jwt.ScopeMember, member.ID), &current); err == nil && current != "" && current != claims.ID {
			return nil, appErrors.ErrInvalidToken
		}
	}

	pair, err := jwt.GenerateTokenPair(member.ID, member.Nickname, jwt.ScopeMember)
	if err != nil {
		return nil, err
	}

	//轮换：拉黑旧刷新令牌，更新访问/刷新会话标记（Redis 不可用时静默降级）
	cfg := config.Get().JWT
	_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, claims.ID), jwt.ReasonRotated, cfg.RefreshTTL())
	_ = s.redisRepo.Set(ctx, jwt.MemberSessionKey(member.ID), pair.AccessJTI, cfg.AccessTTL())
	_ = s.redisRepo.Set(ctx, jwt.RefreshSessionKey(jwt.ScopeMember, member.ID), pair.RefreshJTI, cfg.RefreshTTL())

	return &response.RefreshResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        pair.AccessExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
	}, nil
}

// tokenError 将 JWT 解析错误转换为业务错误码
func tokenError(err error) error {
	if errors.Is(err, jwt.ErrTokenExpired) {
		return appErrors.ErrTokenExpired
	}
	return appErrors.ErrInvalidToken
}

// isBlacklisted 判断 jti 是否命中黑名单（Redis 不可用时 fail-open）
func (s *authService) isBlacklisted(ctx context.Context, scope, jti string) bool {
	if !s.redisRepo.Available(ctx) {
		return false
	}
	var reason string
	return s.redisRepo.Get(ctx, jwt.BlacklistKey(scope, jti), &reason) == nil
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
	member, err := s.memberRepo.FindActiveByPhone(ctx, req.Phone)
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

	//新手机号已注册则拒改（仅看在用的 status<>2；含与原号相同，原号查在用的手机号必然命中自己）
	exist, err := s.memberRepo.FindActiveByPhone(ctx, req.NewPhone)
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

// startSession 单端会话标记：拉黑上一会话的访问/刷新令牌（顶号）后写入新会话 jti；
// Redis 异常静默忽略，不影响登录
func (s *authService) startSession(ctx context.Context, userID uint, accessJTI, refreshJTI string) {
	cfg := config.Get().JWT

	accessKey := jwt.MemberSessionKey(userID)
	var oldAccessJTI string
	if err := s.redisRepo.Get(ctx, accessKey, &oldAccessJTI); err == nil && oldAccessJTI != "" && oldAccessJTI != accessJTI {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, oldAccessJTI), jwt.ReasonKicked, cfg.BlacklistTTL())
	}
	_ = s.redisRepo.Set(ctx, accessKey, accessJTI, cfg.AccessTTL())

	refreshKey := jwt.RefreshSessionKey(jwt.ScopeMember, userID)
	var oldRefreshJTI string
	if err := s.redisRepo.Get(ctx, refreshKey, &oldRefreshJTI); err == nil && oldRefreshJTI != "" && oldRefreshJTI != refreshJTI {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, oldRefreshJTI), jwt.ReasonKicked, cfg.RefreshTTL())
	}
	_ = s.redisRepo.Set(ctx, refreshKey, refreshJTI, cfg.RefreshTTL())
}

// Logout 退出登录：拉黑当前访问令牌与刷新令牌，并清掉会话标记（幂等，Redis 不可用时忽略）
func (s *authService) Logout(ctx context.Context, userID uint, jti string) error {
	cfg := config.Get().JWT
	_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, jti), jwt.ReasonLogout, cfg.BlacklistTTL())

	// 仅当会话标记仍是该 jti（或会话已不存在）时才撤销刷新令牌，避免误杀新登录的会话
	accessKey := jwt.MemberSessionKey(userID)
	var cur string
	if err := s.redisRepo.Get(ctx, accessKey, &cur); err == nil {
		if cur != jti {
			return nil
		}
		_ = s.redisRepo.Del(ctx, accessKey)
	}

	refreshKey := jwt.RefreshSessionKey(jwt.ScopeMember, userID)
	var refreshJTI string
	if err := s.redisRepo.Get(ctx, refreshKey, &refreshJTI); err == nil && refreshJTI != "" {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, refreshJTI), jwt.ReasonLogout, cfg.RefreshTTL())
		_ = s.redisRepo.Del(ctx, refreshKey)
	}
	return nil
}

// revokeSession 注销账号/强制下线时撤销当前访问令牌与刷新令牌
func (s *authService) revokeSession(ctx context.Context, userID uint) {
	cfg := config.Get().JWT

	accessKey := jwt.MemberSessionKey(userID)
	var accessJTI string
	if err := s.redisRepo.Get(ctx, accessKey, &accessJTI); err == nil && accessJTI != "" {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, accessJTI), jwt.ReasonLogout, cfg.BlacklistTTL())
		_ = s.redisRepo.Del(ctx, accessKey)
	}

	refreshKey := jwt.RefreshSessionKey(jwt.ScopeMember, userID)
	var refreshJTI string
	if err := s.redisRepo.Get(ctx, refreshKey, &refreshJTI); err == nil && refreshJTI != "" {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeMember, refreshJTI), jwt.ReasonLogout, cfg.RefreshTTL())
		_ = s.redisRepo.Del(ctx, refreshKey)
	}
}

// Cancel 注销账号：密码 + 短信验证码双重校验，成功后 status=2 并删除登录态
func (s *authService) Cancel(ctx context.Context, userID uint, req *request.CancelRequest) (*response.MessageResponse, error) {
	// 1. 手机号格式
	if err := sms.ValidatorPhone(req.Phone); err != nil {
		return nil, err
	}

	// 2. 取当前账号
	member, err := s.memberRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. 手机号归属校验：必须本人手机号（防拿他人手机号的验证码注销自己）
	if req.Phone != member.Phone {
		return nil, appErrors.ErrPhoneNotMatch
	}

	// 4. 验证码（复用登录/注册的 sms 机制，场景 cancel-account，fail closed）
	if err := sms.Receive(req.Phone, req.Code, sms.SceneCancelAccount); err != nil {
		return nil, appErrors.ErrSmsCodeInvalid
	}

	// 5. 密码
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password)); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	// 6. 置为已注销（数据全部保留，只改状态）
	if err := s.memberRepo.UpdateStatus(ctx, userID, 2); err != nil {
		return nil, err
	}

	// 7. 拉黑当前访问/刷新令牌并清会话标记（MemberAuth 查黑名单，拉黑后立即失效；Redis 异常忽略）
	s.revokeSession(ctx, userID)

	// 8. 返回
	return &response.MessageResponse{Message: "注销成功"}, nil
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
