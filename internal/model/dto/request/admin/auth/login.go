package auth

// LoginRequest 登录请求
type LoginRequest struct {
	UserName  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
	CaptchaId string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
