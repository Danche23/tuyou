package api

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由
type Router struct {
	// ========== 在这里添加你的 Controller 字段 ==========
	// userCtrl  *user.Controller
	// authCtrl  *auth.Controller
}

// NewRouter 创建路由
func NewRouter(
// ========== 在这里添加你的 Service 参数（依赖注入） ==========
// userService service.UserService,
// authService service.AuthService,
) *Router {
	return &Router{
		// userCtrl: user.NewController(userService),
		// authCtrl: auth.NewController(authService),
	}
}

// Setup 设置路由
func (r *Router) Setup(engine *gin.Engine) {
	// 全局中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.CORS())

	// 健康检查
	engine.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "TuYou API is running",
		})
	})

	// API 路由组
	apiGroup := engine.Group("/api")
	{
		_ = apiGroup
		// ========== 在这里注册你的路由 ==========
		// r.authCtrl.RegisterRoutes(apiGroup)
		// r.userCtrl.RegisterRoutes(apiGroup)
	}
}

// Close 关闭所有路由连接
func (r *Router) Close() error {
	return nil
}
