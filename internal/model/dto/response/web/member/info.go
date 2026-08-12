package member

// MemberBaseInfo 会员基础信息（id/昵称/头像/角色）
type MemberBaseInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}
