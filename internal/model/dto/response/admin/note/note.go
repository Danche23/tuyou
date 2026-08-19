package note

// NoteListItem 后台笔记列表项
type NoteListItem struct {
	ID           string   `json:"id"`
	CoverImg     string   `json:"cover_img"`
	Title        string   `json:"title"`
	Nickname     string   `json:"nickname"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	PublishTime  string   `json:"publish_time"`
	ViewCount    int64    `json:"view_count"`
	LikeCount    int64    `json:"like_count"`
	CollectCount int64    `json:"collect_count"`
	CommentCount int64    `json:"comment_count"`
	Heat         int64    `json:"heat"`
	IsSlide      string   `json:"is_slide"`
	Status       string   `json:"status"`
	IsSensitive  int      `json:"is_sensitive"` // 发布时命中敏感词：1=是（内容仍为原文）
}

// ListNoteResp 后台笔记分页响应
type ListNoteResp struct {
	List      []NoteListItem `json:"list"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	Size      int            `json:"size"`
	TotalPage int            `json:"total_page"`
}

// NoteDetailResp 后台笔记详情响应（= 列表项 + content + delete_time）
type NoteDetailResp struct {
	NoteListItem
	Content    string `json:"content"`     // 笔记正文
	DeleteTime string `json:"delete_time"` // 删除时间，未删除时为空字符串
}
