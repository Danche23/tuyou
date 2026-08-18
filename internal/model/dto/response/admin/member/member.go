package member

// ListMemberResp 列表响应
type ListMemberResp struct {
	Total     int64            `json:"total"`
	Page      int              `json:"page"`
	Size      int              `json:"size"`
	TotalPage int              `json:"total_page"`
	List      []MemberListItem `json:"list"`
}

// MemberListItem 会员列表项
type MemberListItem struct {
	ID           string `json:"id"`
	Avatar       string `json:"avatar"`
	Nickname     string `json:"nickname"`
	Phone        string `json:"phone"`
	Status       string `json:"status"` // "启用" / "禁用"
	Bio          string `json:"bio"`
	BgImg        string `json:"bg_img"`
	Role         string `json:"role"`
	RegisterTime string `json:"register_time"`
	LastLogin    string `json:"last_login"`
}

// MemberDetail 会员详情
type MemberDetail struct {
	ID            string `json:"id"`
	Avatar        string `json:"avatar"`
	BgImg         string `json:"bg_img"`
	Nickname      string `json:"nickname"`
	Phone         string `json:"phone"`
	Bio           string `json:"bio"`
	Role          string `json:"role"`
	Status        string `json:"status"` // "启用" / "禁用"
	RegisterTime  string `json:"register_time"`
	LastLogin     string `json:"last_login"`
	FollowCount   int64  `json:"follow_count"`
	FavoriteCount int64  `json:"favorite_count"`
	PublishCount  int64  `json:"publish_count"`
	FansCount     int64  `json:"fans_count"`
}
