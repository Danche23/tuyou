package auth

// AdminInfo 管理员所有信息
type AdminInfo struct {
	ID        string      `json:"id"`
	Username  string      `json:"username"`
	Adminname string      `json:"adminname"`
	Role      string      `json:"role"`
	Avatar    string      `json:"avatar"`
	Status    string      `json:"status"`
	MenuTree  []*MenuTree `json:"menu_tree"`
	PermKeys  []string    `json:"perm_keys"`
}

// MenuTree 菜单树节点
type MenuTree struct {
	ID         string      `json:"id"`
	MenuName   string      `json:"menu_name"`
	ParentID   string      `json:"parent_id"`
	MenuType   int         `json:"menu_type"`
	RoutePath  string      `json:"route_path"`
	Component  string      `json:"component"`
	Icon       string      `json:"icon"`
	HideInMenu int         `json:"hide_in_menu"` // 0显示，1隐藏
	Sort       int         `json:"sort"`
	Children   []*MenuTree `json:"children"`
}

// LoginResp 登录响应
type LoginResp struct {
	// Token 兼容旧客户端，值等同于 access_token
	Token            string      `json:"token"`
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	TokenType        string      `json:"token_type"`
	ExpiresIn        int64       `json:"expires_in"`
	RefreshExpiresIn int64       `json:"refresh_expires_in"`
	ID               string      `json:"id"`
	Username         string      `json:"username"`
	Role             string      `json:"role"`
	Avatar           string      `json:"avatar"`
	MenuTree         []*MenuTree `json:"menu_tree"`
	PermKeys         []string    `json:"perm_keys"`
}

// RefreshResp 刷新令牌响应
type RefreshResp struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

// CaptchaResp 图形验证码响应
type CaptchaResp struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
}
