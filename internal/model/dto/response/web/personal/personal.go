package personal

// Profile 个人主页信息
type Profile struct {
	ID           string `json:"id"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Bio          string `json:"bio"`
	BgImg        string `json:"bgImg"`
	PublishCount int64  `json:"publishCount"`
	FollowCount  int64  `json:"followCount"`
	FansCount    int64  `json:"fansCount"`
}

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

type FollowToggleResponse struct {
	IsFollowed bool `json:"isFollowed"`
}

// AuthorInfo 作者简要信息
type AuthorInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// StrategyCard 攻略卡片（个人中心的点赞/收藏/我的攻略列表）
type StrategyCard struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Author    *AuthorInfo `json:"author"`
	CoverImg  string      `json:"coverImg"`
	IsVideo   bool        `json:"isVideo"` // 是否视频内容
	Title     string      `json:"title"`
	Tag       string      `json:"tag"`
	LikeCount int64       `json:"likeCount"`
	IsLiked   bool        `json:"isLiked"`
}

// NoteCard 笔记卡片（个人中心的点赞/收藏/我的笔记列表）
type NoteCard struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Author    *AuthorInfo `json:"author"`
	CoverImg  string      `json:"coverImg"`
	IsVideo   bool        `json:"isVideo"` // 是否视频内容
	Title     string      `json:"title"`
	Tag       string      `json:"tag"`
	LikeCount int64       `json:"likeCount"`
	IsLiked   bool        `json:"isLiked"`
}
