package apipermissions

import (
	"gorm.io/gorm"
	"strconv"

	"github.com/gin-gonic/gin"
	request "tuyou/internal/model/dto/request/admin/permission/apipermissions"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/permission/apipermissions"
	"tuyou/pkg/response"
)

// Controller 实现控制器
type Controller struct {
	apiPermSvc service.ApiPermService
	redisRepo  repository.RedisRepository
	db         *gorm.DB
}

func NewController(apiPermSvc service.ApiPermService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{apiPermSvc: apiPermSvc, redisRepo: redisRepo, db: db}
}

// List 查询接口权限列表
func (c *Controller) List(ctx *gin.Context) {
	var req request.ListApiPermsReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.apiPermSvc.ListApiPerms(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Modules 获取所有业务模块
func (c *Controller) Modules(ctx *gin.Context) {
	resp, err := c.apiPermSvc.GetModules(ctx)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Insert 新增接口权限
func (c *Controller) Insert(ctx *gin.Context) {
	var req request.InsertApiPermReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.apiPermSvc.InsertApiPerm(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Update 编辑接口权限
func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	var req request.UpdateApiPermReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.apiPermSvc.UpdateApiPerm(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "更新成功", nil)
}

// Delete 删除接口权限
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	if err := c.apiPermSvc.DeleteApiPerm(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "删除成功", nil)
}
