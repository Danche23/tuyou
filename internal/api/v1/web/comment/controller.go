package comment

// 前台评论模块 —— 控制器层（HTTP 入口）
//
// 整个评论功能在项目里的分层是固定的：
//
//	Controller（本文件）→ Service（业务逻辑）→ Repository（SQL）→ MySQL
//	     │                                                        │
//	     └── 只做：取登录用户 → 绑参 → 调 Service → 包装响应 ──────┘
//
// 本文件的每个 Handler 都是"薄封装"，模式完全一样，4 步：
//  1. middleware.GetUserID 拿当前登录用户 ID（未登录返回 401）
//  2. ShouldBindJSON / ShouldBindQuery 把前端参数绑定到 request DTO
//  3. 调 service 层对应方法，拿结果或错误
//  4. response.Success / response.BizError 统一包装
//
// 细粒度权限（"评论者本人或作品作者才能删"）由 service 层判断，
// controller 不碰任何业务规则，只负责 HTTP 层面的收和发。
//
// 本模块共 5 个接口（见 routes.go）：
//  发表评论 / 评论列表 / 删除评论 / 点赞·取消点赞 / 二级评论列表
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
