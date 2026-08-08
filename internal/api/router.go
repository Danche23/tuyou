package api

import (
	handlerAdmin "tuyou/internal/api/v1/admin/auth"
	"tuyou/internal/api/v1/admin/center"
	guideAdmin "tuyou/internal/api/v1/admin/guide"
	"tuyou/internal/api/v1/admin/member"
	noteAdmin "tuyou/internal/api/v1/admin/note"
	adminPermCtrl "tuyou/internal/api/v1/admin/permission/admins"
	adminApiPermCtrl "tuyou/internal/api/v1/admin/permission/apipermissions"
	adminMenuCtrl "tuyou/internal/api/v1/admin/permission/menus"
	uploadCtrl "tuyou/internal/api/v1/upload"
	adminRoleCtrl "tuyou/internal/api/v1/admin/permission/roles"
	"tuyou/internal/api/v1/web/activity"
	handlerAuth "tuyou/internal/api/v1/web/auth"
	"tuyou/internal/api/v1/web/guide"
	homeCtrl "tuyou/internal/api/v1/web/home"
	memberCtrl "tuyou/internal/api/v1/web/member"
	"tuyou/internal/api/v1/web/note"
	tagCtrl "tuyou/internal/api/v1/web/tag"
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由
type Router struct {
	authHandler      *handlerAuth.AuthHandler
	adminHandler     *handlerAdmin.AdminHandler
	centerCtrl       *center.Controller
	noteCtrl         *note.Controller
	memberCtrl       *member.Controller
	adminPerCtrl     *adminPermCtrl.Controller
	adminMenuCtrl    *adminMenuCtrl.Controller
	adminApiPermCtrl *adminApiPermCtrl.Controller
	adminRoleCtrl    *adminRoleCtrl.Controller
	guideCtrl        *guide.Controller
	activityCtrl     *activity.Controller
	noteAdminCtrl    *noteAdmin.Controller
	guideAdminCtrl   *guideAdmin.Controller
	tagCtrl          *tagCtrl.Controller
	webMemberCtrl    *memberCtrl.Controller
	homeCtrl         *homeCtrl.Controller
	uploadCtrl       *uploadCtrl.Controller
}

// NewRouter 创建路由
func NewRouter(
	noteCtrl *note.Controller,
	authHandler *handlerAuth.AuthHandler,
	adminHandler *handlerAdmin.AdminHandler,
	memberCtrl *member.Controller,
	adminPerCtrl *adminPermCtrl.Controller,
	adminMenuCtrl *adminMenuCtrl.Controller,
	adminApiPermCtrl *adminApiPermCtrl.Controller,
	adminRoleCtrl *adminRoleCtrl.Controller,
	guideCtrl *guide.Controller,
	centerCtrl *center.Controller,
	activityCtrl *activity.Controller,
	noteAdminCtrl *noteAdmin.Controller,
	guideAdminCtrl *guideAdmin.Controller,
	tagCtrl *tagCtrl.Controller,
	webMemberCtrl *memberCtrl.Controller,
	homeCtrl *homeCtrl.Controller,
	uploadCtrl *uploadCtrl.Controller,
) *Router {
	return &Router{
		noteCtrl:         noteCtrl,
		authHandler:      authHandler,
		adminHandler:     adminHandler,
		memberCtrl:       memberCtrl,
		adminPerCtrl:     adminPerCtrl,
		adminMenuCtrl:    adminMenuCtrl,
		adminApiPermCtrl: adminApiPermCtrl,
		adminRoleCtrl:    adminRoleCtrl,
		guideCtrl:        guideCtrl,
		activityCtrl:     activityCtrl,
		centerCtrl:       centerCtrl,
		noteAdminCtrl:    noteAdminCtrl,
		guideAdminCtrl:   guideAdminCtrl,
		tagCtrl:          tagCtrl,
		webMemberCtrl:    webMemberCtrl,
		homeCtrl:         homeCtrl,
		uploadCtrl:       uploadCtrl,
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
		r.authHandler.RegisterRoutes(apiGroup)
		r.noteCtrl.RegisterRoutes(apiGroup)
		r.adminHandler.RegisterRoutes(apiGroup)
		r.memberCtrl.RegisterRoutes(apiGroup)
		r.adminPerCtrl.RegisterRoutes(apiGroup)
		r.adminMenuCtrl.RegisterRoutes(apiGroup)
		r.adminApiPermCtrl.RegisterRoutes(apiGroup)
		r.adminRoleCtrl.RegisterRoutes(apiGroup)
		r.guideCtrl.RegisterRoutes(apiGroup)
		r.activityCtrl.RegisterRoutes(apiGroup)
		r.centerCtrl.RegisterRoutes(apiGroup)
		r.noteAdminCtrl.RegisterRoutes(apiGroup)
		r.guideAdminCtrl.RegisterRoutes(apiGroup)
		r.tagCtrl.RegisterRoutes(apiGroup)

		// Web 端会员路由
		r.webMemberCtrl.RegisterRoutes(apiGroup)

		// Web 端社区首页路由
		r.homeCtrl.RegisterRoutes(apiGroup)

		// 文件上传路由
		r.uploadCtrl.RegisterRoutes(apiGroup)
	}
}

// Close 关闭所有路由连接
func (r *Router) Close() error {
	return nil
}
