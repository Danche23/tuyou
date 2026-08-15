package comment

// 前台评论模块 —— 请求 DTO（前端 → 后端的入参契约）
//
// DTO（Data Transfer Object）是分层之间的"传参盒子"：
// 字段上的 binding tag 是 gin 的参数校验规则，如 required=必传、max=300=最长300字，
// 校验不通过时 gin 会直接拒绝请求，业务代码里无需再判。
// 字段的 json:"xxx" / form:"xxx" 决定前端传参的名字：
//  POST 用 JSON 的走 json tag，GET 用 query string 的走 form tag。
//
// 两级评论模型（本项目的核心设计）：
//   - 一级评论：ReplyToID 为空。挂在某个笔记/攻略下。
//   - 二级评论：ReplyToID 指向"归属的一级评论"。无论回复的是一级还是二级，
//     统一都落成二级，且 reply_to_id 永远指向根（一级）评论。
//
// 所以 CreateCommentRequest 里：
//   ReplyToID 为空 → 发一级评论；非空 → 发二级评论。

// CreateCommentRequest 发表评论请求
type CreateCommentRequest struct {
	TargetType      int    `json:"targetType" binding:"required"`      // 目标类型：1=笔记，2=攻略
	TargetID        uint   `json:"targetId" binding:"required"`        // 目标（笔记/攻略）ID
	Content         string `json:"content" binding:"required,max=300"` // 评论内容，最多300字
	ReplyToID       *uint  `json:"replyToId"`                          // 被回复的一级评论ID；空=一级评论，非空=二级评论（回复）
	ReplyToMemberID *uint  `json:"replyToMemberId"`                    // 被回复的用户ID；回复一级评论时可省略（默认取该一级评论作者）
}

// ListCommentRequest 评论列表查询参数
type ListCommentRequest struct {
	TargetType int  `form:"targetType" binding:"required"`           // 目标类型：1=笔记，2=攻略
	TargetID   uint `form:"targetId" binding:"required"`             // 目标ID
	Page       int  `form:"page" binding:"required,min=1,max=10000"` // 页码，从1开始，最大10000
	Size       int  `form:"size" binding:"required,min=1,max=100"`   // 每页大小，最大100
}

// ListRepliesRequest 二级评论列表查询参数
type ListRepliesRequest struct {
	Page int `form:"page" binding:"omitempty,min=1,max=10000"` // 页码，默认1，最大10000
	Size int `form:"size" binding:"omitempty,min=1,max=100"`   // 每页大小，默认10，最大100
}
