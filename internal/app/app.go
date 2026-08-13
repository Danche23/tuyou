package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tuyou/internal/task"

	"tuyou/internal/api"
	activityAdminCtrl "tuyou/internal/api/v1/admin/activity"
	handlerAdmin "tuyou/internal/api/v1/admin/auth"
	centerCtrl "tuyou/internal/api/v1/admin/center"
	guideAdminCtrl "tuyou/internal/api/v1/admin/guide"
	memberCtrl "tuyou/internal/api/v1/admin/member"
	noteAdminCtrl "tuyou/internal/api/v1/admin/note"
	adminPermCtrl "tuyou/internal/api/v1/admin/permission/admins"
	apiPermCtrl "tuyou/internal/api/v1/admin/permission/apipermissions"
	menuCtrl "tuyou/internal/api/v1/admin/permission/menus"
	roleCtrl "tuyou/internal/api/v1/admin/permission/roles"
	workbenchCtrl "tuyou/internal/api/v1/admin/workbench"
	uploadCtrl "tuyou/internal/api/v1/upload"
	"tuyou/internal/api/v1/web/activity"
	handlerAuth "tuyou/internal/api/v1/web/auth"
	commentCtrl "tuyou/internal/api/v1/web/comment"
	"tuyou/internal/api/v1/web/guide"
	homeCtrl "tuyou/internal/api/v1/web/home"
	webMemberCtrl "tuyou/internal/api/v1/web/member"
	"tuyou/internal/api/v1/web/note"
	"tuyou/internal/api/v1/web/personal"
	tagCtrl "tuyou/internal/api/v1/web/tag"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	repoAdminActivity "tuyou/internal/repository/admin/activity"
	repoAdmin "tuyou/internal/repository/admin/auth"
	repoAdminCenter "tuyou/internal/repository/admin/center"
	repoAdminGuide "tuyou/internal/repository/admin/guide"
	repoAdminMember "tuyou/internal/repository/admin/member"
	repoAdminNote "tuyou/internal/repository/admin/note"
	repoAdminPerm "tuyou/internal/repository/admin/permission/admins"
	repoAdminApiPerm "tuyou/internal/repository/admin/permission/apipermissions"
	repoAdminMenu "tuyou/internal/repository/admin/permission/menus"
	repoAdminRole "tuyou/internal/repository/admin/permission/roles"
	repoAdminWorkbench "tuyou/internal/repository/admin/workbench"
	repoActivity "tuyou/internal/repository/web/activity"
	"tuyou/internal/repository/web/auth"
	repoComment "tuyou/internal/repository/web/comment"
	repoGuide "tuyou/internal/repository/web/guide"
	repoHome "tuyou/internal/repository/web/home"
	repoNote "tuyou/internal/repository/web/note"
	repoPersonal "tuyou/internal/repository/web/personal"
	repoTag "tuyou/internal/repository/web/tag"
	serviceAdminActivity "tuyou/internal/service/admin/activity"
	serviceAdmin "tuyou/internal/service/admin/auth"
	serviceAdminCenter "tuyou/internal/service/admin/center"
	serviceAdminGuide "tuyou/internal/service/admin/guide"
	serviceAdminMember "tuyou/internal/service/admin/member"
	serviceAdminNote "tuyou/internal/service/admin/note"
	serviceAdminPerm "tuyou/internal/service/admin/permission/admins"
	serviceAdminApiPerm "tuyou/internal/service/admin/permission/apipermissions"
	serviceAdminMenu "tuyou/internal/service/admin/permission/menus"
	serviceAdminRole "tuyou/internal/service/admin/permission/roles"
	serviceAdminWorkbench "tuyou/internal/service/admin/workbench"
	serviceUpload "tuyou/internal/service/upload"
	serviceActivity "tuyou/internal/service/web/activity"
	serviceAuth "tuyou/internal/service/web/auth"
	serviceComment "tuyou/internal/service/web/comment"
	serviceGuide "tuyou/internal/service/web/guide"
	serviceHome "tuyou/internal/service/web/home"
	serviceNote "tuyou/internal/service/web/note"
	servicePersonal "tuyou/internal/service/web/personal"
	serviceTag "tuyou/internal/service/web/tag"
	"tuyou/internal/storage"
	"tuyou/internal/stream"
	"tuyou/pkg/config"
	"tuyou/pkg/database"
	"tuyou/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用结构体
type App struct {
	cfg               *config.Config
	mysqlDB           *gorm.DB
	redis             *redis.Client
	router            *api.Router
	server            *http.Server
	consumer          *stream.Consumer
	heatTicker        *time.Ticker
	commentHeatTicker *time.Ticker
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{}
}

// Initialize 初始化应用
func (a *App) Initialize(configPath string) error {
	// 1. 加载配置
	if err := a.initConfig(configPath); err != nil {
		return err
	}

	// 2. 初始化日志
	if err := a.initLogger(); err != nil {
		return err
	}

	// 3. 初始化数据库
	if err := a.initDatabase(); err != nil {
		return err
	}

	// 4. 初始化依赖
	a.initDependencies()

	// 5. 初始化路由
	a.initRouter()

	// 6. 初始化服务器
	a.initServer()

	return nil
}

// initConfig 加载配置
func (a *App) initConfig(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	a.cfg = cfg
	return nil
}

// initLogger 初始化日志
func (a *App) initLogger() error {
	if err := logger.Init(&a.cfg.Log); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}

	// 打印启动横幅
	logger.Info("=========================================")
	logger.Info(fmt.Sprintf("欢迎使用 %s", a.cfg.App.Name))
	logger.Info(fmt.Sprintf("版本: %s", a.cfg.App.Version))
	logger.Info(fmt.Sprintf("模式: %s", a.cfg.App.Mode))
	logger.Info("配置加载成功")
	logger.Info("=========================================")

	return nil
}

// initDatabase 初始化数据库
func (a *App) initDatabase() error {
	// 初始化 MySQL
	mysqlDB, err := database.InitMySQL(&a.cfg.Database.MySQL)
	if err != nil {
		return fmt.Errorf("MySQL 初始化失败: %w", err)
	}
	a.mysqlDB = mysqlDB

	// 自动迁移数据库表
	logger.Info("开始数据库迁移...")
	if err := a.mysqlDB.AutoMigrate(
		// 用户模块（2张）
		&entity.Member{},
		&entity.SysAdmin{},
		// 权限模块（5张）
		&entity.SysRole{},
		&entity.SysMenu{},
		&entity.SysPermission{},
		&entity.SysRoleMenu{},
		&entity.SysRolePerm{},
		// 活动模块（2张）
		&entity.Activity{},
		&entity.ActivityJoin{},
		// 内容模块（11张）
		&entity.Note{},
		&entity.Strategy{},
		&entity.Tag{},
		&entity.NoteTag{},
		&entity.StrategyTag{},
		&entity.NoteLike{},
		&entity.NoteCollect{},
		&entity.StrategyLike{},
		&entity.StrategyCollect{},
		&entity.NoteImage{},
		&entity.StrategyImage{},
		// 消息通知模块（3张）
		&entity.MsgTemplate{},
		&entity.MsgInfo{},
		&entity.MsgUserRecord{},
		// 社交模块（3张）
		&entity.Comment{},
		&entity.CommentLike{},
		&entity.Follow{},
	); err != nil {
		logger.Warn("数据库迁移警告", zap.Error(err))
	} else {
		logger.Info("数据库迁移完成")
	}

	// 初始化 Redis（可选，失败不影响核心功能）
	rs, err := database.InitRedis(&a.cfg.Database.Redis)
	if err != nil {
		logger.Warn("Redis 初始化失败，将不影响核心功能", zap.Error(err))
	}
	a.redis = rs

	return nil
}

// initDependencies 初始化依赖注入
func (a *App) initDependencies() {
	// ========== 创建 Redis 仓储 ==========
	redisRepo := repository.NewRedisRepository(a.redis)

	// ========== 创建存储（OSS）==========
	store, err := storage.New(a.cfg.Storage)
	if err != nil {
		logger.Fatal("存储初始化失败", zap.Error(err))
	}

	// ========== 创建 Repository ==========
	memberRepo := auth.NewMemberRepository(a.mysqlDB)
	adminRepo := repoAdmin.NewAdminRepository(a.mysqlDB)
	adminMemberRepo := repoAdminMember.NewMemberRepository(a.mysqlDB)
	adminCenterRepo := repoAdminCenter.NewCenterRepository(a.mysqlDB)
	adminPermRepo := repoAdminPerm.NewAdminsRepository(a.mysqlDB)
	menuRepo := repoAdminMenu.NewMenuRepository(a.mysqlDB)
	apiPermRepo := repoAdminApiPerm.NewApiPermRepository(a.mysqlDB)
	roleRepo := repoAdminRole.NewRoleRepository(a.mysqlDB)
	adminNoteRepo := repoAdminNote.NewNoteRepository(a.mysqlDB)
	adminGuideRepo := repoAdminGuide.NewGuideRepository(a.mysqlDB)
	adminActivityRepo := repoAdminActivity.NewActivityRepository(a.mysqlDB)
	noteRepo := repoNote.NewNoteRepository(a.mysqlDB, redisRepo)
	guideRepo := repoGuide.NewGuideRepository(a.mysqlDB, redisRepo)
	activityRepo := repoActivity.NewActivityRepository(a.mysqlDB)
	personalRepo := repoPersonal.NewPersonalRepository(a.mysqlDB)
	commentRepo := repoComment.NewCommentRepository(a.mysqlDB)
	tagRepo := repoTag.NewTagRepository(a.mysqlDB)
	homeRepo := repoHome.NewHomeRepository(a.mysqlDB)
	adminWorkbenchRepo := repoAdminWorkbench.NewWorkbenchRepository(a.mysqlDB)

	// ========== 创建 Service ==========
	authSvc := serviceAuth.NewAuthService(memberRepo, redisRepo)
	adminAuthSvc := serviceAdmin.NewAdminAuthService(adminRepo, redisRepo)
	noteSvc := serviceNote.NewNoteService(noteRepo, a.mysqlDB)
	adminMemberSvc := serviceAdminMember.NewMemberService(adminMemberRepo)
	adminCenterSvc := serviceAdminCenter.NewCenterService(adminCenterRepo, adminRepo)
	adminPermSvc := serviceAdminPerm.NewAdminService(adminPermRepo)
	menuSvc := serviceAdminMenu.NewMenuService(menuRepo)
	apiPermSvc := serviceAdminApiPerm.NewApiPermService(apiPermRepo)
	roleSvc := serviceAdminRole.NewRoleService(roleRepo)
	adminNoteSvc := serviceAdminNote.NewNoteService(adminNoteRepo, a.mysqlDB)
	adminGuideSvc := serviceAdminGuide.NewGuideService(adminGuideRepo, a.mysqlDB)
	adminActivitySvc := serviceAdminActivity.NewActivityService(adminActivityRepo, activityRepo, a.mysqlDB)
	guideSvc := serviceGuide.NewGuideService(guideRepo, a.mysqlDB)
	activitySvc := serviceActivity.NewActivityService(activityRepo, a.mysqlDB)
	personalSvc := servicePersonal.NewPersonalService(a.mysqlDB, personalRepo, noteRepo, guideRepo, activityRepo)
	commentSvc := serviceComment.NewCommentService(a.mysqlDB, commentRepo)
	tagSvc := serviceTag.NewTagService(tagRepo)
	homeSvc := serviceHome.NewHomeService(homeRepo, noteRepo, guideRepo)
	uploadSvc := serviceUpload.NewUploadService(store)
	adminWorkbenchSvc := serviceAdminWorkbench.NewWorkbenchService(adminWorkbenchRepo)

	// ========== 创建 Handler ==========
	authHandler := handlerAuth.NewAuthHandler(authSvc, redisRepo)
	adminAuthHandler := handlerAdmin.NewAuthHandler(adminAuthSvc, redisRepo)
	personalCtrl := personal.NewController(personalSvc, redisRepo)
	noteCtrl := note.NewController(noteSvc, redisRepo)
	commentC := commentCtrl.NewController(commentSvc, redisRepo)
	memberC := memberCtrl.NewController(adminMemberSvc, redisRepo)
	centerC := centerCtrl.NewController(adminCenterSvc, redisRepo)
	adminPermC := adminPermCtrl.NewController(adminPermSvc, redisRepo)
	menuC := menuCtrl.NewController(menuSvc, redisRepo)
	apiPermC := apiPermCtrl.NewController(apiPermSvc, redisRepo)
	roleC := roleCtrl.NewController(roleSvc, redisRepo)
	noteAdminC := noteAdminCtrl.NewController(adminNoteSvc)
	guideAdminC := guideAdminCtrl.NewController(adminGuideSvc)
	activityAdminC := activityAdminCtrl.NewController(adminActivitySvc)
	guideCtrl := guide.NewController(guideSvc, redisRepo)
	activityCtrl := activity.NewController(activitySvc, redisRepo)
	tagC := tagCtrl.NewController(tagSvc, redisRepo)
	webMemberC := webMemberCtrl.NewController(authSvc, redisRepo)
	homeC := homeCtrl.NewController(homeSvc, redisRepo)
	uploadC := uploadCtrl.NewController(uploadSvc, redisRepo)
	workbenchC := workbenchCtrl.NewController(adminWorkbenchSvc, redisRepo)

	// ========== 创建 Router ==========
	a.router = api.NewRouter(noteCtrl, personalCtrl, commentC, authHandler, adminAuthHandler, memberC, adminPermC, menuC, apiPermC, roleC, guideCtrl, centerC, activityCtrl, noteAdminC, guideAdminC, activityAdminC, tagC, webMemberC, homeC, uploadC, workbenchC)

	// ========== 启动浏览量消费者 ==========
	a.consumer = stream.NewConsumer(redisRepo, a.mysqlDB)
	a.consumer.Start()
	// ========== 启动热度重算定时任务 ==========
	a.heatTicker = task.StartHeatJob(a.mysqlDB)
	// ========== 启动评论热度/评论数校准定时任务 ==========
	a.commentHeatTicker = task.StartCommentHeatJob(a.mysqlDB)
}

// initRouter 初始化路由
func (a *App) initRouter() {
	// 设置 Gin 模式
	gin.SetMode(a.cfg.App.Mode)
}

// initServer 初始化 HTTP 服务器
func (a *App) initServer() {
	engine := gin.New()

	// 注册路由
	a.router.Setup(engine)

	// 创建 HTTP 服务器
	a.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", a.cfg.App.Port),
		Handler:        engine,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

// Run 运行应用
func (a *App) Run() {
	// 启动 HTTP 服务器
	go func() {
		logger.Info("HTTP 服务器启动",
			zap.String("addr", a.server.Addr),
			zap.String("mode", a.cfg.App.Mode),
		)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务器启动失败", zap.Error(err))
		}
	}()

	// 优雅关闭
	a.gracefulShutdown()
}

// gracefulShutdown 优雅关闭
func (a *App) gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
	}

	// 停止后台消费者（在关闭 DB/Redis 之前）
	if a.consumer != nil {
		a.consumer.Stop()
	}

	// 停止热度重算定时任务
	if a.heatTicker != nil {
		a.heatTicker.Stop()
	}

	// 停止评论热度/评论数校准定时任务
	if a.commentHeatTicker != nil {
		a.commentHeatTicker.Stop()
	}

	// 关闭路由连接
	if a.router != nil {
		if err := a.router.Close(); err != nil {
			logger.Error("关闭路由连接失败", zap.Error(err))
		}
	}

	// 关闭数据库连接
	_ = database.CloseMySQL()
	_ = database.CloseRedis()

	// 同步日志
	_ = logger.Sync()

	logger.Info("服务器已关闭")
	logger.Info("=========================================")
}
