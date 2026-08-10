package personal

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册个人中心路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	personal := r.Group("/personal")
	personal.Use(middleware.Auth())
	{
		personal.GET("/follow", c.GetFollowedMembers)
		personal.GET("/fans", c.GetFansMembers)
		personal.POST("/follow/:followeeId", c.ToggleFollow)
		personal.GET("/notes", c.GetMyNotes)
		personal.GET("/liked-strategys", c.GetLikedStrategys)
		personal.GET("/favorite-note", c.GetFavoriteNotes)
		personal.GET("/favorite-strategy", c.GetFavoriteStrategys)
		personal.GET("/liked-notes", c.GetLikedNotes)
		personal.GET("/strategys", c.GetMyStrategys)
		personal.GET("/publish-activity", c.GetPublishActivities)
		personal.GET("/join-activity", c.GetJoinActivities)
	}
}
