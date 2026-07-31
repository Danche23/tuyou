package auth

// AdminInfo 管理员公开信息
type AdminInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Avatar   string `json:"avatar"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string    `json:"token"`
	Admin AdminInfo `json:"admin"`
}

// CaptchaResponse 图形验证码响应
type CaptchaResponse struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
}
