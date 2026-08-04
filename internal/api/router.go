package api

import (
	"tuyou/internal/api/v1/web/activity"
	"tuyou/internal/api/v1/web/guide"
	"tuyou/internal/api/v1/web/note"
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由
type Router struct {
	noteCtrl     *note.Controller
	guideCtrl    *guide.Controller
	activityCtrl *activity.Controller
}

// NewRouter 创建路由
func NewRouter(noteCtrl *note.Controller, guideCtrl *guide.Controller, activityCtrl *activity.Controller) *Router {
	return &Router{
		noteCtrl:     noteCtrl,
		guideCtrl:    guideCtrl,
		activityCtrl: activityCtrl,
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
	apiGroup := engine.Group("/api/v1")
	{
		r.noteCtrl.RegisterRoutes(apiGroup)
		r.guideCtrl.RegisterRoutes(apiGroup)
		r.activityCtrl.RegisterRoutes(apiGroup)
	}
}

// Close 关闭所有路由连接
func (r *Router) Close() error {
	return nil
}
