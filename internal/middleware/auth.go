package middleware

import (
	"errors"
	"strings"

	"tuyou/internal/repository"
	"tuyou/pkg/jwt"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	// ContextUserID 用户 ID 上下文键
	ContextUserID = "user_id"
	// ContextUsername 用户名 上下文键
	ContextUsername = "username"
)

// Auth JWT 认证中间件
// 传 redisRepo 时启用登录态白名单校验（后台接口使用，多端登录 + fail-open）；
// 不传时仅校验 JWT（会员接口使用）。
func Auth(redisRepo ...repository.RedisRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请提供认证令牌")
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "令牌格式错误")
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 白名单校验（仅当传入 redisRepo 时启用）
		if len(redisRepo) > 0 && redisRepo[0] != nil {
			repo := redisRepo[0]
			// fail-open：Redis 不可用时跳过白名单，退化为纯 JWT 校验
			if repo.Available(c.Request.Context()) {
				var v int
				if err := repo.Get(c.Request.Context(), jwt.AdminTokenKey(claims.GetUserID(), parts[1]), &v); err != nil {
					if errors.Is(err, redis.Nil) {
						// 白名单中不存在 → 已退出登录或未登录
						response.Unauthorized(c, "登录已失效，请重新登录")
						c.Abort()
						return
					}
					// 其他异常（Redis 抖动）→ fail-open 放行
				}
			}
		}

		// 将用户信息存入上下文
		c.Set(ContextUserID, claims.GetUserID())
		c.Set(ContextUsername, claims.GetUsername())

		c.Next()
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(ContextUserID); exists {
		return userID.(uint)
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(ContextUsername); exists {
		return username.(string)
	}
	return ""
}

// OptionalAuth 可选的 JWT 认证中间件
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err == nil {
			c.Set(ContextUserID, claims.GetUserID())
			c.Set(ContextUsername, claims.GetUsername())
		}

		c.Next()
	}
}
