package note

// AuthorInfo 作者简要信息
type AuthorInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// CreateNoteResponse 发布笔记响应
type CreateNoteResponse struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Author       *AuthorInfo `json:"author"`
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
	Status       int         `json:"status"`
	PublishTime  string      `json:"publishTime"`
}

// NoteDetailResponse 笔记详情响应
type NoteDetailResponse struct {
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
	Status       int         `json:"status"`
	PublishTime  string      `json:"publishTime"`
}
