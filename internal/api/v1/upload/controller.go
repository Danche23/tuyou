package upload

import (
	"net/http"

	"tuyou/internal/middleware"
	"tuyou/internal/repository"
	serviceUpload "tuyou/internal/service/upload"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 文件上传控制器
type Controller struct {
	uploadSvc serviceUpload.UploadService
	redisRepo repository.RedisRepository
}

// NewController 创建控制器实例
func NewController(uploadSvc serviceUpload.UploadService, redisRepo repository.RedisRepository) *Controller {
	return &Controller{uploadSvc: uploadSvc, redisRepo: redisRepo}
}

// UploadImages 图片批量上传
// POST /api/v1/upload/images （multipart/form-data，字段 files：1~5 张）
func (c *Controller) UploadImages(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: 请选择要上传的图片")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		response.Error(ctx, http.StatusBadRequest, "参数错误: 请选择要上传的图片")
		return
	}

	resp, err := c.uploadSvc.UploadImages(ctx, files)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// UploadVideo 视频上传
// POST /api/v1/upload/video （multipart/form-data，字段 file：单个）
func (c *Controller) UploadVideo(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: 请选择要上传的视频")
		return
	}

	resp, err := c.uploadSvc.UploadVideo(ctx, file)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}
