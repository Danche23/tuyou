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

// GenerateToken 生成 JWT Token，返回 (token, jti, err)。
// jti 每次签发唯一，用于单端登录的会话标记与黑名单；scope 标记身份类型（admin/member），
// 中间件据此隔离前后台接口，防止跨端使用 token。
func GenerateToken(userID uint, username, scope string) (string, string, error) {
	cfg := config.Get().JWT
	jti := newJTI()

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		Scope:    scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.ExpireHours)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tuyou",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.Secret))
	return signed, jti, err
}

// newJTI 生成随机 jti（16 字节随机 hex，32 字符）
func newJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*CustomClaims, error) {
	cfg := config.Get().JWT

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

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

// RefreshToken 刷新 Token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	token, _, err := GenerateToken(claims.UserID, claims.Username, claims.Scope)
	return token, err
}

// 登录态黑名单拉黑原因
const (
	ReasonKicked = "kicked" // 被顶号（其他设备登录）
	ReasonLogout = "logout" // 主动退出 / 被禁用删除
)

// MemberSessionKey 前台当前会话 jti 键（仅登录/退出/禁用时读写，校验不查）
func MemberSessionKey(userID uint) string {
	return fmt.Sprintf("member:session:%d", userID)
}

// AdminSessionKey 后台当前会话 jti 键
func AdminSessionKey(adminID uint) string {
	return fmt.Sprintf("admin:session:%d", adminID)
}

// BlacklistKey 登录态黑名单键（scope：member/admin；jti 每次签发唯一，直接用做 key）
func BlacklistKey(scope, jti string) string {
	return fmt.Sprintf("%s:blacklist:%s", scope, jti)
}
