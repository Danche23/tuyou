package center

// UpdateAvatarRequest 更新头像请求
type UpdateAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required"`
}

// UpdateProfileRequest 编辑账号请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
}

// NotificationListRequest 个人消息列表分页请求
type NotificationListRequest struct {
	Page int `form:"page"` // 页码，从1开始，默认1
	Size int `form:"size"` // 每页大小，默认10，最大100
}
