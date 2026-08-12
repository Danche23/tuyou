package comment

import (
	"net/http"
	"strconv"

	"tuyou/internal/middleware"
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

// ToggleLike 点赞/取消点赞
// POST /api/v1/comments/:id/like
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
