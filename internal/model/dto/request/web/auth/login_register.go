package auth

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone           string `json:"phone" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
	Nickname        string `json:"nickname"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone"    binding:"required"`
	Password string `json:"password" `
	Code     string `json:"code"`
	Type     string `json:"type"`
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Scene string `json:"scene" binding:"required,oneof=register login reset-pwd"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Phone           string `json:"phone" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=NewPassword"`
}
