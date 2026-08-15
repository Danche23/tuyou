package roles

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
		admin.GET("/roles", c.List)
		admin.POST("/roles", c.Insert)
		admin.PUT("/roles/:id", c.Update)
		admin.DELETE("/roles/:id", c.Delete)
		admin.GET("/roles/:id/menus", c.GetMenuTree)
		admin.GET("/roles/:id/permissions", c.GetPermTree)
		admin.PUT("/roles/:id/permissions", c.AssignPermissions)
	}
}
