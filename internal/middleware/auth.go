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

// UploadAuth 前后台上传接口统一认证中间件
// 同时兼容两类登录态（任一有效即放行）：
//   - 后台管理员：`admin:token:{userID}:{md5}` 存在即有效（多端登录，按 token 维度）；
//   - 前台会员：`member:token:{userID}` 的值必须等于当前 token（单端登录，按用户维度）。
//
// 注意顺序：先查管理员白名单，避免后台 token 被数值撞车的前台会员 key 误判为「已在其他设备登
// 注意顺序：先查管理员白名单，避免后台 token 被数值撞车的前台会员 key 误判为「已在其他设备登录」。
// Redis 不可用时 fail-open，退化为纯 JWT 校验。
func UploadAuth(redisRepo repository.RedisRepository) gin.HandlerFunc {
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

		// Redis 可用时才做白名单校验（fail-open）
		if redisRepo != nil && redisRepo.Available(c.Request.Context()) {
			userID := claims.GetUserID()

			// ① 后台管理员白名单：按 token 维度存储，key 存在即有效（多端登录）
			var adminVal int
			valid := redisRepo.Get(c.Request.Context(), jwt.AdminTokenKey(userID, parts[1]), &adminVal) == nil

			// ② 前台会员白名单：按用户维度存最新 token，值必须与当前 token 一致（单端登录）
			if !valid {
				var stored string
				memberErr := redisRepo.Get(c.Request.Context(), jwt.MemberTokenKey(userID), &stored)
				if memberErr == nil && stored == parts[1] {
					valid = true
				} else if memberErr == nil && stored != parts[1] {
					// 会员 key 存在但值是别的 token → 已被顶号/在其他设备登录
					response.Unauthorized(c, "账号已在其他设备登录")
					c.Abort()
					return
				}
			}

			if !valid {
				response.Unauthorized(c, "登录已失效，请重新登录")
				c.Abort()
				return
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

// MemberAuth 前台单端登录认证中间件
// 校验规则：
//   - 校验 JWT 签名；
//   - Redis 可用时，校验当前 token 是否等于该用户白名单中最新 token（`member:token:{userID}`）；
//     白名单中没有 / 不匹配（已登出或被顶号）→ 401；
//   - Redis 不可用时 fail-open，退化为纯 JWT 校验。
func MemberAuth(redisRepo repository.RedisRepository) gin.HandlerFunc {
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

		// 单端白名单校验（fail-open：Redis 不可用时跳过，退化为纯 JWT 校验）
		if redisRepo != nil && redisRepo.Available(c.Request.Context()) {
			var stored string
			if err := redisRepo.Get(c.Request.Context(), jwt.MemberTokenKey(claims.GetUserID()), &stored); err != nil {
				// 白名单中不存在 → 已退出登录或未登录
				response.Unauthorized(c, "登录已失效，请重新登录")
				c.Abort()
				return
			}
			if stored != parts[1] {
				// 不是最新 token → 已在其他设备登录
				response.Unauthorized(c, "账号已在其他设备登录")
				c.Abort()
				return
			}
		}

		// 将用户信息存入上下文
		c.Set(ContextUserID, claims.GetUserID())
		c.Set(ContextUsername, claims.GetUsername())

		c.Next()
	}
}
