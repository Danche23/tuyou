package personal

// UpdateProfileRequest 编辑个人信息请求
type UpdateProfileRequest struct {
	Avatar   string `json:"avatar"   binding:"required"`
	Nickname string `json:"nickname" binding:"required,max=20"`
	Bio      string `json:"bio"      binding:"required,max=100"`
	BgImg    string `json:"bgImg"    binding:"required"`
}
