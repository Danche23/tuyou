package personal

// MemberCard 他人公开资料卡（用户详情，GET /personal/member/:id）
// 不含 phone/password；status：0=禁用 1=正常 2=已注销（已注销仍返回资料卡）
type MemberCard struct {
	ID            string `json:"id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Bio           string `json:"bio"`
	BgImg         string `json:"bgImg"`
	Role          string `json:"role"`
	Status        int    `json:"status"`
	FollowCount   int64  `json:"followCount"`
	FansCount     int64  `json:"fansCount"`
	PublishCount  int64  `json:"publishCount"`
	FavoriteCount int64  `json:"favoriteCount"`
	LeadingCount  int64  `json:"leadingCount"`
	IsFollowed    bool   `json:"isFollowed"` // 我是否关注 TA
	IsFollower    bool   `json:"isFollower"` // TA 是否关注我（互关判断）
}

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
