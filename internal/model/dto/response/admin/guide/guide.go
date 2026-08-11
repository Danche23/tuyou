package guide

// GuideListItem 后台攻略列表项
type GuideListItem struct {
	ID           string   `json:"id"`
	CoverImg     string   `json:"cover_img"`
	Title        string   `json:"title"`
	Nickname     string   `json:"nickname"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	Difficulty   string   `json:"difficulty"`
	PublishTime  string   `json:"publish_time"`
	ViewCount    int64    `json:"view_count"`
	LikeCount    int64    `json:"like_count"`
	CollectCount int64    `json:"collect_count"`
	CommentCount int64    `json:"comment_count"`
	Heat         int64    `json:"heat"`
	IsSlide      string   `json:"is_slide"`
	Status       string   `json:"status"`
}

// ListGuideResp 后台攻略分页响应
type ListGuideResp struct {
	List      []GuideListItem `json:"list"`
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	Size      int             `json:"size"`
	TotalPage int             `json:"total_page"`
}

// GuideDetailResp 后台攻略详情响应（= 列表项 + content + distance/duration/transport/video_url + delete_time）
type GuideDetailResp struct {
	GuideListItem
	Content    string `json:"content"`     // 攻略正文
	Distance   string `json:"distance"`    // 距离
	Duration   string `json:"duration"`    // 总时长
	Transport  string `json:"transport"`   // 交通方式（用户自由填写）
	VideoURL   string `json:"video_url"`   // 视频地址
	DeleteTime string `json:"delete_time"` // 删除时间，未删除时为空字符串
}
