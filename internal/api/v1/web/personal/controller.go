package personal

import (
	"net/http"
	"strconv"
	"tuyou/internal/middleware"
	requestPersonal "tuyou/internal/model/dto/request/web/personal"
	servicePersonal "tuyou/internal/service/web/personal"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 个人中心控制器
type Controller struct {
	personalSvc servicePersonal.PersonalService
}

// NewController 创建控制器实例
func NewController(personalSvc servicePersonal.PersonalService) *Controller {
	return &Controller{personalSvc: personalSvc}
}

func (c *Controller) GetProfile(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	profile, err := c.personalSvc.GetProfile(memberID)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, profile)
}

// UpdateProfile 编辑个人信息
// PUT /api/v1/personal
func (c *Controller) UpdateProfile(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req requestPersonal.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}

	profile, err := c.personalSvc.UpdateProfile(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, profile)
}

// GetFollowedMembers 我关注的用户列表
// GET /api/v1/personal/follow
func (c *Controller) GetFollowedMembers(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}

	list, total, err := c.personalSvc.GetFollowedMembers(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(list, total, req.Page, req.Size))
}

// 粉丝列表
func (c *Controller) GetFansMembers(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}

	list, total, err := c.personalSvc.GetFansMembers(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(list, total, req.Page, req.Size))
}

// ToggleFollow 关注/取关
// POST /api/v1/personal/follow/:followeeId
func (c *Controller) ToggleFollow(ctx *gin.Context) {
	followerID := middleware.GetUserID(ctx)
	if followerID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	followeeID, err := strconv.ParseUint(ctx.Param("followeeId"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "参数错误: followeeId格式不正确")
		return
	}
	resp, err := c.personalSvc.ToggleFollow(followerID, uint(followeeID))
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, resp)
}

// 我的笔记
func (c *Controller) GetMyNotes(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	cards, total, err := c.personalSvc.GetMyNotes(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(cards, total, req.Page, req.Size))
}

// 点赞的攻略
func (c *Controller) GetLikedStrategys(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	cards, total, err := c.personalSvc.GetLikedStrategys(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(cards, total, req.Page, req.Size))
}

// GetFavoriteNotes 我收藏的笔记
// GET /api/v1/personal/favorite-note
func (c *Controller) GetFavoriteNotes(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	cards, total, err := c.personalSvc.GetFavoriteNotes(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(cards, total, req.Page, req.Size))
}

// GetFavoriteStrategys 我收藏的攻略
// GET /api/v1/personal/favorite-strategy
func (c *Controller) GetFavoriteStrategys(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	cards, total, err := c.personalSvc.GetFavoriteStrategys(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(cards, total, req.Page, req.Size))
}

// GetLikedNotes 我点赞的笔记
// GET /api/v1/personal/liked-notes
func (c *Controller) GetLikedNotes(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	cards, total, err := c.personalSvc.GetLikedNotes(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(cards, total, req.Page, req.Size))
}

// GetMyStrategys 我发布的攻略
// GET /api/v1/personal/strategys
func (c *Controller) GetMyStrategys(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	cards, total, err := c.personalSvc.GetMyStrategys(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(cards, total, req.Page, req.Size))
}

// GetPublishActivities 我发布的活动
// GET /api/v1/personal/publish-activity
func (c *Controller) GetPublishActivities(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	list, total, err := c.personalSvc.GetPublishActivities(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(list, total, req.Page, req.Size))
}

// GetJoinActivities 我参加的活动
// GET /api/v1/personal/join-activity
func (c *Controller) GetJoinActivities(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}
	var req response.PageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "参数校验失败:"+err.Error())
		return
	}
	list, total, err := c.personalSvc.GetJoinActivities(memberID, req.Page, req.Size)
	if err != nil {
		response.BizError(ctx, err)
		return
	}
	response.Success(ctx, response.NewPageResponse(list, total, req.Page, req.Size))
}
