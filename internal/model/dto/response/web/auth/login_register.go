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
	// Token 兼容旧客户端，值等同于 access_token
	Token            string     `json:"token"`
	AccessToken      string     `json:"access_token"`
	RefreshToken     string     `json:"refresh_token"`
	TokenType        string     `json:"token_type"`
	ExpiresIn        int64      `json:"expires_in"`
	RefreshExpiresIn int64      `json:"refresh_expires_in"`
	Member           MemberInfo `json:"member"`
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
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
