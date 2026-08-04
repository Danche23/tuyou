package center

// UpdateAvatarRequest 更新头像请求
type UpdateAvatarRequest struct {
	Avatar string `json:"avatar" binding:"required"`
}

// UpdateProfileRequest 编辑账号请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
}
