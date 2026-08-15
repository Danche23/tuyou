package menus

import (
	"github.com/gin-gonic/gin"
	"tuyou/internal/middleware"
)

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	admin.Use(
		middleware.Auth(c.redisRepo),
		middleware.PermissionAuth(c.db),
	)
	{
		admin.GET("/menus", c.List)
		admin.GET("/menus/all", c.ListAll)
		admin.POST("/menus", c.Insert)
		admin.PUT("/menus/:id", c.Update)
		admin.DELETE("/menus/:id", c.Delete)
	}
}
