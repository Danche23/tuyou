package guide

// AuthorInfo 作者简要信息
type AuthorInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// RouteInfo 攻略路线信息
type RouteInfo struct {
	Distance   string `json:"distance"`
	Duration   string `json:"duration"`
	Difficulty string `json:"difficulty"`
	Transport  string `json:"transport"`
}

// CreateGuideResponse 发布攻略响应（简洁：仅返回新内容 ID 和类型）
type CreateGuideResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// CollectToggleResponse 攻略收藏/取消收藏响应
type CollectToggleResponse struct {
	IsFavorited bool `json:"isFavorited"`
}

// LikeToggleResponse 攻略点赞/取消点赞响应
type LikeToggleResponse struct {
	IsLiked   bool  `json:"isLiked"`
	LikeCount int64 `json:"likeCount"`
}

// GuideDetailResponse 攻略详情响应
type GuideDetailResponse struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Author       *AuthorInfo `json:"author"`
	IsFollowed   bool        `json:"isFollowed"`
	Title        string      `json:"title"`
	CoverImg     string      `json:"coverImg"`
	Images       []string    `json:"images"`
	VideoURL     string      `json:"videoUrl"`
	Content      string      `json:"content"`
	ViewCount    int64       `json:"viewCount"`
	LikeCount    int64       `json:"likeCount"`
	CollectCount int64       `json:"collectCount"`
	CommentCount int64       `json:"commentCount"`
	IsLiked      bool        `json:"isLiked"`
	IsFavorited  bool        `json:"isFavorited"`
	Category     int         `json:"category"`
	Tags         []string    `json:"tags"`
	PublishTime  string      `json:"publishTime"`
	RouteInfo    *RouteInfo  `json:"routeInfo"`
}
