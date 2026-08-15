package workbench

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/workbench"
	"tuyou/pkg/response"
)

// Controller 后台工作台控制器
type Controller struct {
	workbenchSvc service.WorkbenchService
	redisRepo    repository.RedisRepository
	db           *gorm.DB
}

// NewController 创建控制器实例
func NewController(workbenchSvc service.WorkbenchService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{workbenchSvc: workbenchSvc, redisRepo: redisRepo, db: db}
}

// GetWorkbench 获取工作台统计数据
func (c *Controller) GetWorkbench(ctx *gin.Context) {
	resp, err := c.workbenchSvc.GetWorkbench(ctx)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
