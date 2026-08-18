package middleware

import (
	"gorm.io/gorm"
	"net/http"
	"strings"

	"tuyou/internal/repository"
	"tuyou/pkg/jwt"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID 用户 ID 上下文键
	ContextUserID = "user_id"
	// ContextUsername 用户名 上下文键
	ContextUsername = "username"
)

// blacklistAuth 单端黑名单校验核心逻辑：
//   - 校验 JWT 签名；
//   - Redis 可用时，校验 token 的 jti 是否命中 scopes 对应端黑名单，命中 → 401；
//   - Redis 不可用时 fail-open，退化为纯 JWT 校验。
//
// 会话标记（session）不参与校验，仅在登录/退出/禁用时读写。
func blacklistAuth(c *gin.Context, redisRepo repository.RedisRepository, scopes ...string) {
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

	// 黑名单校验（fail-open：Redis 不可用时跳过，退化为纯 JWT 校验）
	if redisRepo != nil && redisRepo.Available(c.Request.Context()) {
		for _, scope := range scopes {
			var reason string
			if err := redisRepo.Get(c.Request.Context(), jwt.BlacklistKey(scope, claims.ID), &reason); err == nil {
				response.Unauthorized(c, blacklistMessage(reason))
				c.Abort()
				return
			}
		}
	}

	// 将用户信息存入上下文
	c.Set(ContextUserID, claims.GetUserID())
	c.Set(ContextUsername, claims.GetUsername())

	c.Next()
}

// blacklistMessage 黑名单原因 → 提示文案
func blacklistMessage(reason string) string {
	switch reason {
	case jwt.ReasonKicked:
		return "账号已在其他设备登录"
	default:
		return "登录已失效，请重新登录"
	}
}

// Auth 后台单端登录认证中间件
// 传 redisRepo 时启用黑名单校验（单端登录 + fail-open）；不传时仅校验 JWT。
func Auth(redisRepo ...repository.RedisRepository) gin.HandlerFunc {
	var repo repository.RedisRepository
	if len(redisRepo) > 0 {
		repo = redisRepo[0]
	}
	return func(c *gin.Context) {
		blacklistAuth(c, repo, "admin")
	}
}

// MemberAuth 前台单端登录认证中间件
// 校验 JWT 签名 + 黑名单；Redis 不可用时 fail-open。
func MemberAuth(redisRepo repository.RedisRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		blacklistAuth(c, redisRepo, "member")
	}
}

// PermissionAuth 权限认证，校验用户是否有该权限
func PermissionAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID := GetUserID(c)
		if adminID == 0 {
			response.Error(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}

		method := c.Request.Method
		path := c.FullPath()

		// ① 该接口是否在 sys_permission 全局表登记（status=1 的启用权限）
		var registered int64
		if err := db.Table("sys_permission").
			Where("status = 1 AND request_method = ? AND api_path = ?", method, path).
			Count(&registered).Error; err != nil {
			c.Next() // DB 异常 fail-open
			return
		}
		if registered == 0 {
			c.Next() // 未登记的接口默认放行
			return
		}

		// ② 已登记 → 查当前管理员角色拥有的接口权限，判断是否允许该操作
		var perms []struct {
			RequestMethod string
			APIPath       string
		}
		err := db.Table("sys_permission AS p").
			Select("p.request_method, p.api_path").
			Joins("JOIN sys_role_perm rp ON rp.perm_id = p.id").
			Joins("JOIN sys_admin a ON a.role_id = rp.role_id").
			Where("a.id = ? AND p.status = 1", adminID).
			Scan(&perms).Error
		if err != nil {
			c.Next() // DB 异常 fail-open
			return
		}

		for _, p := range perms {
			if p.RequestMethod == method && p.APIPath == path {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "无访问权限")
		c.Abort()
	}
}

// UploadAuth 前后台上传接口统一认证中间件
// 同时兼容两类登录态：后台管理员 / 前台会员的 token 任一命中黑名单即拒。
// Redis 不可用时 fail-open，退化为纯 JWT 校验。
func UploadAuth(redisRepo repository.RedisRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		blacklistAuth(c, redisRepo, "admin", "member")
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
