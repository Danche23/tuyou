package member

import (
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/member"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/member"
	"tuyou/pkg/response"
)

// Controller 会员管理控制器
type Controller struct {
	memberSvc service.MemberService
	redisRepo repository.RedisRepository
	db        *gorm.DB
}

// NewController 创建控制器实例
func NewController(memberSvc service.MemberService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{memberSvc: memberSvc, redisRepo: redisRepo, db: db}
}

// ListMembers 查询会员列表（/admin/members）
func (c *Controller) ListMembers(ctx *gin.Context) {
	var req request.ListMemberReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	//会员列表同时展示会员(1)和领队(0)：role 传 nil 不按角色过滤
	resp, err := c.memberSvc.ListMembers(ctx, nil, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// ListLeaders 查询领队列表（/admin/leaders）
func (c *Controller) ListLeaders(ctx *gin.Context) {
	var req request.ListMemberReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	//领队的角色用 0 表示
	leaderRole := 0
	resp, err := c.memberSvc.ListMembers(ctx, &leaderRole, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetDetail 获取会员/领队详情
func (c *Controller) GetDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}
	resp, err := c.memberSvc.GetMemberDetail(ctx, uint(id))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Update 编辑会员/领队
func (c *Controller) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}
	var req request.UpdateMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	if err := c.memberSvc.UpdateMember(ctx, uint(id), &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "更新成功", nil)
}

// Delete 删除会员/领队
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}
	if err := c.memberSvc.DeleteMember(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "删除成功", nil)
}

// BatchDelete 批量删除（/admin/members/batch-delete、/admin/leaders/batch-delete 共用）
func (c *Controller) BatchDelete(ctx *gin.Context) {
	var req request.BatchDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	if err := c.memberSvc.BatchDeleteMembers(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "删除成功", nil)
}

// BatchUpdateStatus 批量修改状态（/admin/members/batch-status、/admin/leaders/batch-status 共用）
func (c *Controller) BatchUpdateStatus(ctx *gin.Context) {
	var req request.BatchUpdateStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	if err := c.memberSvc.BatchUpdateStatus(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.SuccessWithMessage(ctx, "更新成功", nil)
}

// ExportMembers 导出会员（/admin/members/export）
func (c *Controller) ExportMembers(ctx *gin.Context) {
	c.export(ctx, 1)
}

// ExportLeaders 导出领队（/admin/leaders/export）
func (c *Controller) ExportLeaders(ctx *gin.Context) {
	c.export(ctx, 0)
}

// export 导出通用逻辑：将该角色全部记录生成为 xlsx 二进制流并下载
func (c *Controller) export(ctx *gin.Context, role int) {
	list, err := c.memberSvc.ExportMembers(ctx, role)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	data, err := buildExcel(list)
	if err != nil {
		response.InternalError(ctx, "导出失败")
		return
	}
	prefix := "member_export_"
	if role == 0 {
		prefix = "leader_export_"
	}
	filename := prefix + time.Now().Format("20060102") + ".xlsx"
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
