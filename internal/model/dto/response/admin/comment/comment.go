package comment

// 后台评论管理模块 —— 响应 DTO（管理端出参契约）
//
// 与前台响应的差异：
//   - 字段全 snake_case（target_type、create_time），对应 postman 文档。
//   - ID 仍是 string（防 JS 精度丢失，跟前台一致）。
//   - 多了 target_title（所属笔记/攻略标题）、status/delete_time（软删状态）。
//
// 关键点：status 不是数据库字段，是"派生值"——由 delete_time 有没有值算出来的：
//   delete_time 为空 → status=正常；非空 → status=已删除。
// 因为后台要连已软删的评论一起管理（可能要点删除、清理），所以列表里能看到"已删除"的记录。
// 整个后台模块复用这一个结构体（列表项和详情是同一个东西）。

// CommentItem 后台评论列表项/详情（列表与详情共用，字段与 postman 一致）
type CommentItem struct {
	ID              string `json:"id"`                // 评论 ID（纯数字）
	Nickname        string `json:"nickname"`          // 评论用户昵称
	Avatar          string `json:"avatar"`            // 评论用户头像
	TargetType      string `json:"target_type"`       // 目标类型：笔记/攻略
	TargetTitle     string `json:"target_title"`      // 所属笔记/攻略标题
	Content         string `json:"content"`           // 评论内容
	Level           string `json:"level"`             // 评论层级：一级/二级
	ReplyToNickname string `json:"reply_to_nickname"` // 被回复用户昵称，一级评论为空
	LikeCount       int64  `json:"like_count"`        // 点赞数
	Heat            int64  `json:"heat"`              // 热度
	CreateTime      string `json:"create_time"`       // 评论创建时间
	Status          string `json:"status"`            // 状态：正常/已删除（派生自 delete_time）
	IsSensitive     int    `json:"is_sensitive"`      // 发布时是否命中敏感词：1=是（内容为原文）
	HitWords        []string `json:"hit_words"`       // 命中的敏感词数组（发布时快照；存储为分号分隔字符串，出参拆成数组）
	DeleteTime      string `json:"delete_time"`       // 删除时间，未删除为空
}

// ListCommentResp 后台评论分页响应
type ListCommentResp struct {
	List      []CommentItem `json:"list"`
	Total     int64         `json:"total"`
	Page      int           `json:"page"`
	Size      int           `json:"size"`
	TotalPage int           `json:"total_page"`
}
