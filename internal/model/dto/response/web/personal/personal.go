package personal

// FollowedMember 关注列表里展示的用户信息
type FollowedMember struct {
	FolloweeID string `json:"followeeId"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Bio        string `json:"bio"`
	IsFollowed bool   `json:"isFollowed"`
}

// FansMember 粉丝列表里展示的用户信息
type FansMember struct {
	FollowerID string `json:"followerId"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Bio        string `json:"bio"`
	IsFollowed bool   `json:"isFollowed"`
}

type Noteview struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type FollowToggleResponse struct {
	IsFollowed bool `json:"isFollowed"`
}
