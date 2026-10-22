package jwt

import "github.com/golang-jwt/jwt/v5"

// CustomClaims 自定义 JWT Claims
type CustomClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Scope     string `json:"scope"`      // 身份类型：admin=后台管理员，member=前台会员
	TokenType string `json:"token_type"` // 令牌类型：access=访问令牌，refresh=刷新令牌
	jwt.RegisteredClaims
}

// GetUserID 获取用户 ID
func (c *CustomClaims) GetUserID() uint {
	return c.UserID
}

// GetUsername 获取用户名
func (c *CustomClaims) GetUsername() string {
	return c.Username
}
