package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"tuyou/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenInvalid = errors.New("token 无效")
	ErrTokenExpired = errors.New("token 已过期")
)

// 登录身份类型（写进 token 的 scope 声明，中间件据此校验端身份）
const (
	ScopeAdmin  = "admin"  // 后台管理员
	ScopeMember = "member" // 前台会员
)

// 令牌类型：访问令牌用于接口认证，刷新令牌仅用于换取新的令牌对
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// TokenPair 一次签发的访问令牌 + 刷新令牌
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessJTI        string
	RefreshJTI       string
	AccessExpiresIn  int64
	RefreshExpiresIn int64
}

// GenerateTokenPair 签发访问令牌与刷新令牌。
// 两个令牌各自持有唯一 jti，便于分别做黑名单与轮换；scope 标记身份类型（admin/member）。
func GenerateTokenPair(userID uint, username, scope string) (*TokenPair, error) {
	cfg := config.Get().JWT
	accessTTL := cfg.AccessTTL()
	refreshTTL := cfg.RefreshTTL()

	accessToken, accessJTI, err := generateToken(userID, username, scope, TokenTypeAccess, accessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshJTI, err := generateToken(userID, username, scope, TokenTypeRefresh, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessJTI:        accessJTI,
		RefreshJTI:       refreshJTI,
		AccessExpiresIn:  int64(accessTTL.Seconds()),
		RefreshExpiresIn: int64(refreshTTL.Seconds()),
	}, nil
}

// GenerateToken 只签发访问令牌，保留给旧调用方使用；新代码请使用 GenerateTokenPair。
func GenerateToken(userID uint, username, scope string) (string, string, error) {
	return generateToken(userID, username, scope, TokenTypeAccess, config.Get().JWT.AccessTTL())
}

// generateToken 生成指定类型、指定有效期的 JWT，返回 (token, jti, err)
func generateToken(userID uint, username, scope, tokenType string, ttl time.Duration) (string, string, error) {
	jti := newJTI()
	now := time.Now()

	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Scope:     scope,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "tuyou",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.Get().JWT.Secret))
	return signed, jti, err
}

// newJTI 生成随机 jti（16 字节随机 hex，32 字符）
func newJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ParseToken 解析访问令牌。
// 刷新令牌不能用于接口认证；旧版未携带 token_type 的令牌按访问令牌兼容处理。
func ParseToken(tokenString string) (*CustomClaims, error) {
	claims, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "" && claims.TokenType != TokenTypeAccess {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// ParseRefreshToken 解析刷新令牌，访问令牌会被拒绝。
func ParseRefreshToken(tokenString string) (*CustomClaims, error) {
	claims, err := parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// ParseRefreshTokenForScope 解析刷新令牌并校验身份端（admin/member），防止跨端刷新。
func ParseRefreshTokenForScope(tokenString, scope string) (*CustomClaims, error) {
	claims, err := ParseRefreshToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Scope != scope {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// parseToken 校验签名并解析 Claims
func parseToken(tokenString string) (*CustomClaims, error) {
	cfg := config.Get().JWT

	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// 登录态黑名单拉黑原因
const (
	ReasonKicked  = "kicked"  // 被顶号（其他设备登录）
	ReasonLogout  = "logout"  // 主动退出 / 被禁用删除
	ReasonRotated = "rotated" // 刷新令牌已轮换
)

// MemberSessionKey 前台当前访问令牌 jti 键（仅登录/退出/禁用时读写，校验不查）
func MemberSessionKey(userID uint) string {
	return fmt.Sprintf("member:session:%d", userID)
}

// AdminSessionKey 后台当前访问令牌 jti 键
func AdminSessionKey(adminID uint) string {
	return fmt.Sprintf("admin:session:%d", adminID)
}

// RefreshSessionKey 当前刷新令牌 jti 键（刷新/登出/禁用时读写，用于刷新令牌轮换）
func RefreshSessionKey(scope string, userID uint) string {
	return fmt.Sprintf("%s:refresh_session:%d", scope, userID)
}

// BlacklistKey 登录态黑名单键（scope：member/admin；jti 每次签发唯一，直接用做 key）
func BlacklistKey(scope, jti string) string {
	return fmt.Sprintf("%s:blacklist:%s", scope, jti)
}
