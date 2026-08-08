package apipermissions

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
)

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(c.redisRepo))
	{
		admin.GET("/api-permissions", c.List)
		admin.POST("/api-permissions", c.Insert)
		admin.PUT("/api-permissions/:id", c.Update)
		admin.DELETE("/api-permissions/:id", c.Delete)
	}
}
