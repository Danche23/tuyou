package sensitiveword

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	request "tuyou/internal/model/dto/request/admin/sensitiveword"
	"tuyou/internal/repository"
	service "tuyou/internal/service/admin/sensitiveword"
	"tuyou/pkg/response"
)

// Controller 后台敏感词管理控制器
type Controller struct {
	wordSvc   service.SensitiveWordService
	redisRepo repository.RedisRepository
	db        *gorm.DB
}

func NewController(wordSvc service.SensitiveWordService, redisRepo repository.RedisRepository, db *gorm.DB) *Controller {
	return &Controller{wordSvc: wordSvc, redisRepo: redisRepo, db: db}
}

// List GET /api/v1/admin/sensitive-words?page=&size=&keyword=
func (c *Controller) List(ctx *gin.Context) {
	var req request.ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	resp, err := c.wordSvc.List(ctx, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// Create POST /api/v1/admin/sensitive-words  body: {"word":"新词"}
func (c *Controller) Create(ctx *gin.Context) {
	var req request.CreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败")
		return
	}
	if err := c.wordSvc.Create(ctx, &req); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// Delete DELETE /api/v1/admin/sensitive-words/:id
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "ID参数不合法")
		return
	}
	if err := c.wordSvc.Delete(ctx, uint(id)); err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
