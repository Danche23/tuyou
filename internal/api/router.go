package api

import (
	activityAdmin "tuyou/internal/api/v1/admin/activity"
	handlerAdmin "tuyou/internal/api/v1/admin/auth"
	"tuyou/internal/api/v1/admin/center"
	commentAdmin "tuyou/internal/api/v1/admin/comment"
	guideAdmin "tuyou/internal/api/v1/admin/guide"
	"tuyou/internal/api/v1/admin/member"
	noteAdmin "tuyou/internal/api/v1/admin/note"
	adminPermCtrl "tuyou/internal/api/v1/admin/permission/admins"
	adminApiPermCtrl "tuyou/internal/api/v1/admin/permission/apipermissions"
	adminMenuCtrl "tuyou/internal/api/v1/admin/permission/menus"
	adminRoleCtrl "tuyou/internal/api/v1/admin/permission/roles"
	tagAdmin "tuyou/internal/api/v1/admin/tag"
	workbenchCtrl "tuyou/internal/api/v1/admin/workbench"
	uploadCtrl "tuyou/internal/api/v1/upload"
	"tuyou/internal/api/v1/web/activity"
	handlerAuth "tuyou/internal/api/v1/web/auth"
	commentCtrl "tuyou/internal/api/v1/web/comment"
	"tuyou/internal/api/v1/web/guide"
	homeCtrl "tuyou/internal/api/v1/web/home"
	memberCtrl "tuyou/internal/api/v1/web/member"
	messageCtrl "tuyou/internal/api/v1/web/message"
	"tuyou/internal/api/v1/web/note"
	"tuyou/internal/api/v1/web/personal"
	searchCtrl "tuyou/internal/api/v1/web/search"
	tagCtrl "tuyou/internal/api/v1/web/tag"
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由
type Router struct {
	authHandler       *handlerAuth.AuthHandler
	adminHandler      *handlerAdmin.AdminHandler
	centerCtrl        *center.Controller
	commentAdminCtrl  *commentAdmin.Controller
	noteCtrl          *note.Controller
	personalCtrl      *personal.Controller
	commentCtrl       *commentCtrl.Controller
	memberCtrl        *member.Controller
	adminPerCtrl      *adminPermCtrl.Controller
	adminMenuCtrl     *adminMenuCtrl.Controller
	adminApiPermCtrl  *adminApiPermCtrl.Controller
	adminRoleCtrl     *adminRoleCtrl.Controller
	guideCtrl         *guide.Controller
	activityCtrl      *activity.Controller
	noteAdminCtrl     *noteAdmin.Controller
	guideAdminCtrl    *guideAdmin.Controller
	activityAdminCtrl *activityAdmin.Controller
	tagAdminCtrl      *tagAdmin.Controller
	tagCtrl           *tagCtrl.Controller
	webMemberCtrl     *memberCtrl.Controller
	homeCtrl          *homeCtrl.Controller
	searchCtrl        *searchCtrl.Controller
	messageCtrl       *messageCtrl.Controller
	uploadCtrl        *uploadCtrl.Controller
	workbenchCtrl     *workbenchCtrl.Controller
}

// NewRouter 创建路由
func NewRouter(
	noteCtrl *note.Controller,
	personalCtrl *personal.Controller,
	commentCtrl *commentCtrl.Controller,
	authHandler *handlerAuth.AuthHandler,
	adminHandler *handlerAdmin.AdminHandler,
	memberCtrl *member.Controller,
	adminPerCtrl *adminPermCtrl.Controller,
	adminMenuCtrl *adminMenuCtrl.Controller,
	adminApiPermCtrl *adminApiPermCtrl.Controller,
	adminRoleCtrl *adminRoleCtrl.Controller,
	guideCtrl *guide.Controller,
	centerCtrl *center.Controller,
	commentAdminCtrl *commentAdmin.Controller,
	activityCtrl *activity.Controller,
	noteAdminCtrl *noteAdmin.Controller,
	guideAdminCtrl *guideAdmin.Controller,
	activityAdminCtrl *activityAdmin.Controller,
	tagAdminCtrl *tagAdmin.Controller,
	tagCtrl *tagCtrl.Controller,
	webMemberCtrl *memberCtrl.Controller,
	homeCtrl *homeCtrl.Controller,
	searchCtrl *searchCtrl.Controller,
	messageCtrl *messageCtrl.Controller,
	uploadCtrl *uploadCtrl.Controller,
	workbenchCtrl *workbenchCtrl.Controller,
) *Router {
	return &Router{
		noteCtrl:          noteCtrl,
		personalCtrl:      personalCtrl,
		commentCtrl:       commentCtrl,
		authHandler:       authHandler,
		adminHandler:      adminHandler,
		memberCtrl:        memberCtrl,
		adminPerCtrl:      adminPerCtrl,
		adminMenuCtrl:     adminMenuCtrl,
		adminApiPermCtrl:  adminApiPermCtrl,
		adminRoleCtrl:     adminRoleCtrl,
		guideCtrl:         guideCtrl,
		activityCtrl:      activityCtrl,
		centerCtrl:        centerCtrl,
		commentAdminCtrl:  commentAdminCtrl,
		noteAdminCtrl:     noteAdminCtrl,
		guideAdminCtrl:    guideAdminCtrl,
		activityAdminCtrl: activityAdminCtrl,
		tagAdminCtrl:      tagAdminCtrl,
		tagCtrl:           tagCtrl,
		webMemberCtrl:     webMemberCtrl,
		homeCtrl:          homeCtrl,
		searchCtrl:        searchCtrl,
		messageCtrl:       messageCtrl,
		uploadCtrl:        uploadCtrl,
		workbenchCtrl:     workbenchCtrl,
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

		// Web 端认证路由
		r.authHandler.RegisterRoutes(apiGroup)

		// Web 端笔记路由
		r.noteCtrl.RegisterRoutes(apiGroup)

		// 后台认证路由
		r.adminHandler.RegisterRoutes(apiGroup)

		//web个人模块路由
		r.personalCtrl.RegisterRoutes(apiGroup)

		// web评论路由（前台，/api/v1/comments，用户视角，MemberAuth 登录）
		r.commentCtrl.RegisterRoutes(apiGroup)

		// 后台会员管理路由
		r.memberCtrl.RegisterRoutes(apiGroup)

		// 后台权限-管理员管理路由
		r.adminPerCtrl.RegisterRoutes(apiGroup)

		// 后台权限-菜单管理路由
		r.adminMenuCtrl.RegisterRoutes(apiGroup)

		// 后台权限-接口权限管理路由
		r.adminApiPermCtrl.RegisterRoutes(apiGroup)

		// 后台权限-角色管理路由
		r.adminRoleCtrl.RegisterRoutes(apiGroup)

		// Web 端攻略路由
		r.guideCtrl.RegisterRoutes(apiGroup)

		// Web 端活动路由
		r.activityCtrl.RegisterRoutes(apiGroup)

		// 后台个人中心路由
		r.centerCtrl.RegisterRoutes(apiGroup)

		// 后台评论管理路由（/api/v1/admin/comments，运营视角，Auth 管理员登录）
		r.commentAdminCtrl.RegisterRoutes(apiGroup)

		// 后台笔记管理路由
		r.noteAdminCtrl.RegisterRoutes(apiGroup)

		// 后台攻略管理路由
		r.guideAdminCtrl.RegisterRoutes(apiGroup)

		// 后台活动管理路由
		r.activityAdminCtrl.RegisterRoutes(apiGroup)

		// 后台标签管理路由
		r.tagAdminCtrl.RegisterRoutes(apiGroup)

		// Web 端标签路由
		r.tagCtrl.RegisterRoutes(apiGroup)

		// Web 端会员路由
		r.webMemberCtrl.RegisterRoutes(apiGroup)

		// Web 端社区首页路由
		r.homeCtrl.RegisterRoutes(apiGroup)

		// Web 端全量搜索路由
		r.searchCtrl.RegisterRoutes(apiGroup)

		// Web 端消息路由
		r.messageCtrl.RegisterRoutes(apiGroup)

		// 文件上传路由
		r.uploadCtrl.RegisterRoutes(apiGroup)

		// 后台工作台路由
		r.workbenchCtrl.RegisterRoutes(apiGroup)
	}
}

// Close 关闭所有路由连接
func (r *Router) Close() error {
	return nil
}
