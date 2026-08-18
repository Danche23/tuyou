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

// MessageResponse 验证码发送、注册和密码重置格式
type MessageResponse struct {
	Message string `json:"message"`
}

// CaptchaResp 图形验证码响应
type CaptchaResp struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
}
