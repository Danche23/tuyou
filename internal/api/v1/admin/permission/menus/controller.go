package menus

import (
	"github.com/gin-gonic/gin"
	"strconv"
	request "tuyou/internal/model/dto/request/admin/permission/menus"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/permission/menus"
	"tuyou/pkg/response"
)

// Controller 实现控制器
type Controller struct {
	menuSvc   service.MenuService
	redisRepo repository.RedisRepository
}

func NewController(menuSvc service.MenuService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{menuSvc: menuSvc, redisRepo: redisRepo}
}

// ListAll 查询全部菜单（父菜单选择器用）
func (c *Controller) ListAll(ctx *gin.Context) {
	resp, err := c.menuSvc.ListAllMenus(ctx)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// List 查询菜单列表
func (c *Controller) List(ctx *gin.Context) {
	var req request.ListMenusReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.menuSvc.ListMenus(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Insert 新增菜单
func (c *Controller) Insert(ctx *gin.Context) {
	var req request.InsertMenuReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.menuSvc.InsertMenu(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Update 编辑菜单
func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	var req request.UpdateMenuReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.menuSvc.UpdateMenu(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "更新成功", nil)
}

// Delete 删除菜单
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误：id格式不正确")
		return
	}

	if err := c.menuSvc.DeleteMenu(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "删除成功", nil)
}
