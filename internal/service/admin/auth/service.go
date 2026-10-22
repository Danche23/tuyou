package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"strconv"
	request "tuyou/internal/model/dto/request/admin/auth"
	response "tuyou/internal/model/dto/response/admin/auth"
	responseCenter "tuyou/internal/model/dto/response/admin/center"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	"tuyou/internal/repository/admin/auth"
	repoCenter "tuyou/internal/repository/admin/center"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/jwt"

	"golang.org/x/crypto/bcrypt"

	"tuyou/internal/service/admin/auth/b64c"
	"tuyou/internal/stream"
	"tuyou/pkg/config"
)

type authService struct {
	adminRepo    auth.AdminRepository
	redisRepo    repository.RedisRepository
	adminMsgRepo repoCenter.AdminMsgRepository
	adminHub     *stream.SSEHub
}

func NewAdminAuthService(adminRepo auth.AdminRepository, redisRepo repository.RedisRepository, adminMsgRepo repoCenter.AdminMsgRepository, adminHub *stream.SSEHub) AdminAuthService {
	return &authService{adminRepo: adminRepo, redisRepo: redisRepo, adminMsgRepo: adminMsgRepo, adminHub: adminHub}
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
func (s *authService) Login(ctx context.Context, req *request.LoginRequest, userAgent string) (*response.LoginResp, error) {

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

	//签发访问令牌 + 刷新令牌
	pair, err := jwt.GenerateTokenPair(admin.ID, admin.Username, jwt.ScopeAdmin)
	if err != nil {
		return nil, err
	}

	//单端黑名单：新登录顶号——拉黑上一会话的访问/刷新令牌后覆盖会话标记（Redis 不可用时忽略，不影响登录）
	s.startSession(ctx, admin.ID, pair.AccessJTI, pair.RefreshJTI)

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

	// 异步生成设备登录提醒（系统通知，写 msg_info + msg_user_record 并推送未读数），best-effort
	go s.notifyLogin(context.Background(), admin.ID, userAgent)

	return &response.LoginResp{
		Token:            pair.AccessToken,
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        pair.AccessExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
		ID:               strconv.Itoa(int(admin.ID)),
		Username:         admin.Username,
		Role:             role,
		Avatar:           admin.Avatar,
		MenuTree:         buildMenuTree(menus),
		PermKeys:         buildPermKeys(perms),
	}, nil
}

// Refresh 使用刷新令牌换取新的令牌对（刷新令牌一次性使用，旧令牌立即拉黑轮换）
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*response.RefreshResp, error) {
	claims, err := jwt.ParseRefreshTokenForScope(refreshToken, jwt.ScopeAdmin)
	if err != nil {
		return nil, tokenError(err)
	}

	if s.isBlacklisted(ctx, jwt.ScopeAdmin, claims.ID) {
		return nil, appErrors.ErrInvalidToken
	}

	//刷新前确认管理员仍然有效，避免已禁用/删除账号继续续期
	admin, err := s.adminRepo.FindByID(ctx, claims.UserID)
	if err != nil || admin.Status != 1 {
		return nil, appErrors.ErrInvalidToken
	}

	//校验当前刷新令牌：Redis 可用时，会话中记录的 jti 必须与本次令牌一致，防止旧令牌重放
	if s.redisRepo.Available(ctx) {
		var current string
		if err := s.redisRepo.Get(ctx, jwt.RefreshSessionKey(jwt.ScopeAdmin, admin.ID), &current); err == nil && current != "" && current != claims.ID {
			return nil, appErrors.ErrInvalidToken
		}
	}

	pair, err := jwt.GenerateTokenPair(admin.ID, admin.Username, jwt.ScopeAdmin)
	if err != nil {
		return nil, err
	}

	//轮换：拉黑旧刷新令牌，更新访问/刷新会话标记（Redis 不可用时静默降级）
	cfg := config.Get().JWT
	_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, claims.ID), jwt.ReasonRotated, cfg.RefreshTTL())
	_ = s.redisRepo.Set(ctx, jwt.AdminSessionKey(admin.ID), pair.AccessJTI, cfg.AccessTTL())
	_ = s.redisRepo.Set(ctx, jwt.RefreshSessionKey(jwt.ScopeAdmin, admin.ID), pair.RefreshJTI, cfg.RefreshTTL())

	return &response.RefreshResp{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        pair.AccessExpiresIn,
		RefreshExpiresIn: pair.RefreshExpiresIn,
	}, nil
}

// startSession 单端会话标记：拉黑上一会话的访问/刷新令牌（顶号）后写入新会话 jti；
// Redis 异常静默忽略，不影响登录
func (s *authService) startSession(ctx context.Context, adminID uint, accessJTI, refreshJTI string) {
	cfg := config.Get().JWT

	accessKey := jwt.AdminSessionKey(adminID)
	var oldAccessJTI string
	if err := s.redisRepo.Get(ctx, accessKey, &oldAccessJTI); err == nil && oldAccessJTI != "" && oldAccessJTI != accessJTI {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, oldAccessJTI), jwt.ReasonKicked, cfg.BlacklistTTL())
	}
	_ = s.redisRepo.Set(ctx, accessKey, accessJTI, cfg.AccessTTL())

	refreshKey := jwt.RefreshSessionKey(jwt.ScopeAdmin, adminID)
	var oldRefreshJTI string
	if err := s.redisRepo.Get(ctx, refreshKey, &oldRefreshJTI); err == nil && oldRefreshJTI != "" && oldRefreshJTI != refreshJTI {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, oldRefreshJTI), jwt.ReasonKicked, cfg.RefreshTTL())
	}
	_ = s.redisRepo.Set(ctx, refreshKey, refreshJTI, cfg.RefreshTTL())
}

// Logout 退出登录：拉黑当前访问令牌与刷新令牌，并清掉会话标记（幂等，Redis 不可用时忽略）
func (s *authService) Logout(ctx context.Context, adminID uint, jti string) error {
	cfg := config.Get().JWT
	_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, jti), jwt.ReasonLogout, cfg.BlacklistTTL())

	// 仅当会话标记仍是该 jti（或会话已不存在）时才撤销刷新令牌，避免误杀新登录的会话
	accessKey := jwt.AdminSessionKey(adminID)
	var cur string
	if err := s.redisRepo.Get(ctx, accessKey, &cur); err == nil {
		if cur != jti {
			return nil
		}
		_ = s.redisRepo.Del(ctx, accessKey)
	}

	refreshKey := jwt.RefreshSessionKey(jwt.ScopeAdmin, adminID)
	var refreshJTI string
	if err := s.redisRepo.Get(ctx, refreshKey, &refreshJTI); err == nil && refreshJTI != "" {
		_ = s.redisRepo.Set(ctx, jwt.BlacklistKey(jwt.ScopeAdmin, refreshJTI), jwt.ReasonLogout, cfg.RefreshTTL())
		_ = s.redisRepo.Del(ctx, refreshKey)
	}
	return nil
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

// notifyLogin 生成设备登录提醒（系统通知）并推送最新未读数，best-effort，失败不影响登录
func (s *authService) notifyLogin(ctx context.Context, adminID uint, userAgent string) {
	content := fmt.Sprintf("您的账号于%s在%s登录，如非本人操作，请尽快修改密码",
		time.Now().Format("2006-01-02 15:04:05"), deviceOf(userAgent))
	if err := s.adminMsgRepo.Create(ctx, adminID, content); err != nil {
		return
	}
	if cnt, err := s.adminMsgRepo.UnreadCount(ctx, adminID); err == nil {
		s.adminHub.Publish(adminID, "unread", &responseCenter.AdminUnreadCount{UnreadCount: cnt})
	}
}

// deviceOf 从 User-Agent 粗略识别设备与浏览器，用于登录提醒文案
func deviceOf(ua string) string {
	os := "未知设备"
	switch {
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "Macintosh"), strings.Contains(ua, "Mac OS"):
		os = "macOS"
	case strings.Contains(ua, "iPhone"), strings.Contains(ua, "iPad"):
		os = "iOS"
	case strings.Contains(ua, "Android"):
		os = "Android"
	}
	browser := "未知浏览器"
	switch {
	case strings.Contains(ua, "Edg"):
		browser = "Edge"
	case strings.Contains(ua, "Chrome"):
		browser = "Chrome"
	case strings.Contains(ua, "Firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "Safari"):
		browser = "Safari"
	}
	return os + " - " + browser
}

// buildMenuTree 将扁平菜单列表组装成树
func buildMenuTree(menus []entity.SysMenu) []*response.MenuTree {
	nodes := make(map[uint]*response.MenuTree, len(menus))
	roots := make([]*response.MenuTree, 0, len(menus))
	for i := range menus {
		m := &menus[i]
		nodes[m.ID] = &response.MenuTree{
			ID:         strconv.Itoa(int(m.ID)),
			MenuName:   m.MenuName,
			ParentID:   strconv.Itoa(int(m.ParentID)),
			MenuType:   m.MenuType,
			RoutePath:  m.RoutePath,
			Component:  m.Component,
			Icon:       m.Icon,
			HideInMenu: m.HideInMenu,
			Sort:       m.Sort,
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

// buildPermKeys 提取权限标识列表
func buildPermKeys(perms []entity.SysPermission) []string {
	keys := make([]string, 0, len(perms))
	for _, p := range perms {
		keys = append(keys, p.PermKey)
	}
	return keys
}

// statusStr 将账号状态映射为展示文案：1启用，0禁用
func statusStr(status int) string {
	if status == 1 {
		return "启用"
	}
	return "禁用"
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
	//查询角色菜单与权限
	menus, err := s.adminRepo.FindMenusByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}
	perms, err := s.adminRepo.FindPermsByRoleID(ctx, admin.RoleID)
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
		MenuTree:  buildMenuTree(menus),
		PermKeys:  buildPermKeys(perms),
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
