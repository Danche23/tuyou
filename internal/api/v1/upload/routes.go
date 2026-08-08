package upload

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册上传路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	up := r.Group("/upload")
	up.Use(middleware.MemberAuth(c.redisRepo))
	{
		up.POST("/images", c.UploadImages)
		up.POST("/video", c.UploadVideo)
	}
}
