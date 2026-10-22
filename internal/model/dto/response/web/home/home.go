package home

// SlideItem 首页轮播项
type SlideItem struct {
	ID            string `json:"id"`                      // 内容ID（字符串，与现有接口一致）
	Type          string `json:"type"`                    // 内容类型：note | guide | activity
	Title         string `json:"title"`                   // 标题
	Cover         string `json:"cover"`                   // 封面图URL
	Tag           string `json:"tag"`                     // 标签/分类（活动取分类，笔记攻略取首个标签）
	SignUpStart   string `json:"signUpStart,omitempty"`   // 报名开始时间（仅 activity）
	CurrentPeople int    `json:"currentPeople,omitempty"` // 当前报名人数（仅 activity）
	MaxPeople     int    `json:"maxPeople,omitempty"`     // 最大人数（仅 activity）
}

// FeedAuthor Feed 作者信息
type FeedAuthor struct {
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像URL
}

// FeedItem Feed 流条目（笔记/攻略统一结构）
type FeedItem struct {
	ID        string      `json:"id"`        // 内容ID（字符串）
	Title     string      `json:"title"`     // 标题
	Images    []string    `json:"images"`    // 封面图URL数组（图文/视频内容均取第一张图）
	VideoURL  *string     `json:"videoUrl"`  // 视频URL（图文内容为 null）
	Tags      []string    `json:"tags"`      // 标签数组（Feed 只填第一个启用标签）
	Member    *FeedAuthor `json:"member"`    // 作者信息
	LikeCount int64       `json:"likeCount"` // 点赞数
	IsLiked   bool        `json:"isLiked"`   // 当前用户是否已点赞
	Type      string      `json:"type"`      // 内容类型：note | guide
	Category  int         `json:"category"`  // 分类ID
}

// FeedResponse Feed 流响应（分页）
//
// 同时提供两套分页信息以兼容新旧前端：
//   - 游标：next_cursor / has_more（推荐使用，不受翻页期间新增内容影响）
//   - 页码：total / total_page / page（保留，值仍为真实统计）
type FeedResponse struct {
	List       []FeedItem `json:"list"`        // 当前页数据
	Total      int64      `json:"total"`       // 总条数
	TotalPage  int        `json:"total_page"`  // 总页数
	Page       int        `json:"page"`        // 当前页码
	Size       int        `json:"size"`        // 每页条数
	NextCursor string     `json:"next_cursor"` // 下一页游标；空串表示已到末页
	HasMore    bool       `json:"has_more"`    // 是否还有下一页
}
