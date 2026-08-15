package personal

// UpdateProfileRequest 编辑个人信息请求
// 部分更新语义：只有非空字段才会被写入；avatar/bio/bgImg 空=不改（可省略），nickname 必填。
type UpdateProfileRequest struct {
	Avatar   string `json:"avatar"   binding:"max=255"`         // 头像：可选，空=不改（DB varchar(255)）
	Nickname string `json:"nickname" binding:"required,max=20"` // 昵称：必填，最长20字
	Bio      string `json:"bio"      binding:"max=100"`         // 简介：可选，空=不改
	BgImg    string `json:"bgImg"    binding:"max=255"`         // 背景图：可选，空=不改（DB varchar(255)）
}
