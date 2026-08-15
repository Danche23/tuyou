package comment

// 后台评论管理模块 —— 控制器层（管理端 HTTP 入口）
//
// 与前台评论（web/comment）是同一张 comment 表，但服务对象不同：
//   - 前台（web）：普通用户视角，只能看未删评论，删=软删。
//   - 后台（admin）：运营/管理员视角，要"看到一切"（含已软删的），删=彻底硬删。
//
// 本模块 4 个接口（见 router.go），模式与前台 Controller 一致，都是薄封装：
//  1. 绑参（query 用 ShouldBindQuery，body 用 ShouldBindJSON）
//  2. 调 service 层
//  3. response.Success / response.BizError 包装
//
// 与前台最不同的两点：
//   - 鉴权用 middleware.Auth（管理员登录态 + 白名单），不是 MemberAuth（用户登录态）。
//   - 列表/详情用 Unscoped() 连"已软删的评论"一起查出来，status 字段标记正常/已删除。
import (
	"gorm.io/gorm"
	"strconv"

	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/comment"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"

	request "tuyou/internal/model/dto/request/admin/comment"
)

// Controller 后台评论管理控制器
type Controller struct {
	commentSvc service.CommentService
	redisRepo  repository.RedisRepository // middleware.Auth 后台登录态白名单校验需要
	db         *gorm.DB
}

// NewController 创建控制器实例
func NewController(commentSvc service.CommentService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{commentSvc: commentSvc, redisRepo: redisRepo, db: db}
}

// ListComments 分页查询评论
// GET /api/v1/admin/comments
func (c *Controller) ListComments(ctx *gin.Context) {
	var req request.ListCommentReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	resp, err := c.commentSvc.ListComments(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// GetCommentDetail 获取评论详情
// GET /api/v1/admin/comments/:id
func (c *Controller) GetCommentDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	resp, err := c.commentSvc.GetCommentDetail(ctx, uint(id))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// DeleteComment 彻底删除评论（一级连带删除其下所有二级）
// DELETE /api/v1/admin/comments/:id
func (c *Controller) DeleteComment(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误: id格式不正确")
		return
	}

	if err := c.commentSvc.DeleteComment(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// BatchDeleteComments 批量彻底删除评论
// POST /api/v1/admin/comments/batch-delete  body: {"ids":["3","5"]}
func (c *Controller) BatchDeleteComments(ctx *gin.Context) {
	var req request.BatchDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}

	if err := c.commentSvc.BatchDeleteComments(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
