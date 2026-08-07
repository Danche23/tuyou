package auth

// AdminInfo 管理员所有信息
type AdminInfo struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Adminname string `json:"adminname"`
	Role      string `json:"role"`
	Avatar    string `json:"avatar"`
	Status    bool   `json:"status"`
}

// MenuTree 菜单树节点
type MenuTree struct {
	ID        string      `json:"id"`
	MenuName  string      `json:"menu_name"`
	ParentID  string      `json:"parent_id"`
	MenuType  int         `json:"menu_type"`
	RoutePath string      `json:"route_path"`
	Sort      int         `json:"sort"`
	Children  []*MenuTree `json:"children"`
}

// PermItem 权限树叶子节点（单个接口权限）
type PermItem struct {
	ID            string `json:"id"`
	PermName      string `json:"perm_name"`
	PermKey       string `json:"perm_key"`
	RequestMethod string `json:"request_method"`
	APIPath       string `json:"api_path"`
}

// PermTree 权限树节点（按业务模块分组）
type PermTree struct {
	Module   string     `json:"module"`
	Children []PermItem `json:"children"`
}

// LoginResp 登录响应
type LoginResp struct {
	Token    string      `json:"token"`
	ID       string      `json:"id"`
	Username string      `json:"username"`
	Role     string      `json:"role"`
	Avatar   string      `json:"avatar"`
	MenuTree []*MenuTree `json:"menu_tree"`
	PermTree []PermTree  `json:"perm_tree"`
}

// CaptchaResp 图形验证码响应
type CaptchaResp struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
}
