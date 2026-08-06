package admins

import (
	"github.com/gin-gonic/gin"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/permission/admins"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/permission/admins"
	"tuyou/pkg/response"
)

// Controller 实现控制器
type Controller struct {
	adminSvc  service.AdminService
	redisRepo repository.RedisRepository
}

func NewController(adminSvc service.AdminService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{adminSvc: adminSvc, redisRepo: redisRepo}
}

// ListRoles 查询角色列表
func (c *Controller) ListRoles(ctx *gin.Context) {
	resp, err := c.adminSvc.ListRoles(ctx)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// List 查询管理员列表
func (c *Controller) List(ctx *gin.Context) {
	var req request.ListAdminsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.adminSvc.ListAdmin(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Insert 新增管理员
func (c *Controller) Insert(ctx *gin.Context) {
	var req request.InsertAdminReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.adminSvc.InsertAdmin(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Update 编辑管理员
func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	var req request.UpdateAdminReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.adminSvc.UpdateAdmin(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "更新成功", nil)
}

// Delete 删除管理员
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	if err := c.adminSvc.DeleteAdmin(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "删除成功", nil)
}

func (c *Controller) ResetPassword(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	var req request.ResetAdminPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.adminSvc.ResetAdminPassword(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "密码更新成功", nil)
}
