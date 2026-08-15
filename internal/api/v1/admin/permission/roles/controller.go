package roles

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/permission/roles"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/permission/roles"
	"tuyou/pkg/response"
)

// Controller 实现控制器
type Controller struct {
	roleSvc   service.RoleService
	redisRepo repository.RedisRepository
	db        *gorm.DB
}

func NewController(roleSvc service.RoleService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{roleSvc: roleSvc, redisRepo: redisRepo, db: db}
}

// List 查询角色列表
func (c *Controller) List(ctx *gin.Context) {
	var req request.ListRolesReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.roleSvc.ListRoles(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Insert 新增角色
func (c *Controller) Insert(ctx *gin.Context) {
	var req request.InsertRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.roleSvc.InsertRole(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Update 编辑角色
func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	var req request.UpdateRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.roleSvc.UpdateRole(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "更新成功", nil)
}

// Delete 删除角色
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	if err := c.roleSvc.DeleteRole(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetMenuTree 获取菜单分配树
func (c *Controller) GetMenuTree(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	resp, err := c.roleSvc.GetMenuTree(ctx, uint(id))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetPermTree 获取权限分配树
func (c *Controller) GetPermTree(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	resp, err := c.roleSvc.GetPermTree(ctx, uint(id))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// AssignPermissions 角色分配菜单与接口权限
func (c *Controller) AssignPermissions(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	var req request.AssignPermissionsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.roleSvc.AssignPermissions(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "分配成功", nil)
}
