package comment

// 前台评论模块 —— 响应 DTO（后端 → 前端的出参契约）
//
// 与请求 DTO 对应，这里是返回给前端的结构。要点：
//   - ID 用 string 而不是 uint：避免大数在前端 JS 里丢失精度（JS 只能安全表示 2^53）。
//   - CommentItem（一级）与 CommentReplyItem（二级）的区别：一级带 replies（热评预览）
//     和 replyCount（回复总数），二级不再嵌套。
//   - replies 里放的"热评预览"最多 2 条，且只在点赞数 >10 时展示——见 service 层的 buildCommentItems。
//   - canDelete 是给前端的"是否显示删除按钮"开关，由 service 判断（本人或作品作者）。
//
// 注意：JSON 字段用的是驼峰（likeCount、createdAt），与后台评论管理的 snake_case 不同，
// 两套是给不同前端用的（web 端 vs 管理端），别混。

// LikeToggleResponse 点赞/取消点赞响应
type LikeToggleResponse struct {
	IsLiked   bool  `json:"isLiked"`   // 本次操作后的点赞状态：true=已赞，false=未赞
	LikeCount int64 `json:"likeCount"` // 本次操作后的最新点赞数
}

// CommentUser 评论用户信息
type CommentUser struct {
	ID       string `json:"id"`       // 用户ID（字符串，防 JS 精度丢失）
	Nickname string `json:"nickname"` // 昵称
	Avatar   string `json:"avatar"`   // 头像URL
}

// CommentItem 评论对象（一级评论列表项 / 发表评论响应）
type CommentItem struct {
	ID              string              `json:"id"`              // 评论ID（字符串）
	User            *CommentUser        `json:"user"`            // 评论用户信息
	Content         string              `json:"content"`         // 评论内容
	LikeCount       int64               `json:"likeCount"`       // 点赞数
	IsLiked         bool                `json:"isLiked"`         // 当前登录用户是否已点赞
	ReplyToID       string              `json:"replyToId"`       // 归属的一级评论ID；一级评论本身为空字符串
	ReplyToNickname string              `json:"replyToNickname"` // 被回复用户昵称；一级评论为空
	CreatedAt       string              `json:"createdAt"`       // 创建时间 yyyy-MM-dd HH:mm:ss
	Replies         []*CommentReplyItem `json:"replies"`         // 热评预览：该一级评论下热度最高前2条二级（点赞数>10才展示，否则空数组）
	ReplyCount      int64               `json:"replyCount"`      // 该一级评论下二级评论总数
	CanDelete       bool                `json:"canDelete"`       // 当前用户能否删除（评论者本人或作品作者=true）
}

// CommentReplyItem 二级评论对象（加载更多返回，无嵌套回复）
type CommentReplyItem struct {
	ID              string       `json:"id"`              // 评论ID（字符串）
	User            *CommentUser `json:"user"`            // 评论用户信息
	Content         string       `json:"content"`         // 评论内容
	LikeCount       int64        `json:"likeCount"`       // 点赞数
	IsLiked         bool         `json:"isLiked"`         // 当前登录用户是否已点赞
	ReplyToID       string       `json:"replyToId"`       // 归属的一级评论ID（即父评论ID）
	ReplyToNickname string       `json:"replyToNickname"` // 被回复用户昵称
	CreatedAt       string       `json:"createdAt"`       // 创建时间 yyyy-MM-dd HH:mm:ss
	CanDelete       bool         `json:"canDelete"`       // 当前用户能否删除（评论者本人或作品作者=true）
}
