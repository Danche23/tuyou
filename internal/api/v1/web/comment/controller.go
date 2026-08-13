package comment

import (
	"net/http"
	"strconv"

	"tuyou/internal/middleware"
	requestComment "tuyou/internal/model/dto/request/web/comment"
	"tuyou/internal/repository"
	serviceComment "tuyou/internal/service/web/comment"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 评论控制器
type Controller struct {
	commentSvc serviceComment.CommentService
	redisRepo  repository.RedisRepository // MemberAuth 中间件需要，note 模块同款
}

// NewController 创建控制器实例
func NewController(commentSvc serviceComment.CommentService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{commentSvc: commentSvc, redisRepo: redisRepo}
}

// Create 发表评论
// POST /api/v1/comments
func (c *Controller) Create(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestComment.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := c.commentSvc.CreateComment(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// List 评论列表
func (c *Controller) List(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestComment.ListCommentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := c.commentSvc.ListComments(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Delete 删除评论（评论者本人或作品作者可删）
// DELETE /api/v1/comments/:id
func (c *Controller) Delete(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	commentID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	if err := c.commentSvc.DeleteComment(uint(commentID), memberID); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ToggleLike 点赞/取消点赞
// POST /api/v1/comments/like/:id
func (c *Controller) ToggleLike(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	commentID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	resp, err := c.commentSvc.ToggleLike(uint(commentID), memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// ListReplies 二级评论列表
// GET /api/v1/comments/replies/:id?page=&size=
func (c *Controller) ListReplies(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	idStr := ctx.Param("id")
	commentID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: id格式不正确")
		return
	}

	var req requestComment.ListRepliesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}
	// size 默认10
	page, size := req.Page, req.Size
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}

	resp, err := c.commentSvc.ListReplies(uint(commentID), memberID, page, size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
