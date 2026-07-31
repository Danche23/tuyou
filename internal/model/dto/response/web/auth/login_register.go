package auth

// MemberInfo 会员公开信息
type MemberInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token  string     `json:"token"`
	Member MemberInfo `json:"member"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Token  string     `json:"token"`
	Member MemberInfo `json:"member"`
}

// MessageResponse 验证码发送和密码重置格式
type MessageResponse struct {
	Message string `json:"message"`
}
