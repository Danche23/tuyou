package jwt

import (
	"crypto/md5"
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

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint, username string) (string, error) {
	cfg := config.Get().JWT

	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.ExpireHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tuyou",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
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

	return GenerateToken(claims.UserID, claims.Username)
}

// AdminTokenKey 后台登录态白名单 Key（按 token 维度存储，支持多端登录）
// 使用 md5 压缩 token，避免把完整 token 拼进 Redis key
func AdminTokenKey(adminID uint, token string) string {
	hash := md5.Sum([]byte(token))
	return fmt.Sprintf("admin:token:%d:%x", adminID, hash)
}

// MemberTokenKey 前台登录态白名单 Key（按用户维度存储，支持单端登录）
// 存的是最新 token，新登录覆盖旧值实现顶号
func MemberTokenKey(userID uint) string {
	return fmt.Sprintf("member:token:%d", userID)
}
