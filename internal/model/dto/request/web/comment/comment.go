package comment

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
	TargetType int  `form:"targetType" binding:"required"`         // 目标类型：1=笔记，2=攻略
	TargetID   uint `form:"targetId" binding:"required"`           // 目标ID
	Page       int  `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size       int  `form:"size" binding:"required,min=1,max=100"` // 每页大小，最大100
}

// ListRepliesRequest 二级评论列表查询参数
type ListRepliesRequest struct {
	Page int `form:"page" binding:"omitempty,min=1"`         // 页码，默认1
	Size int `form:"size" binding:"omitempty,min=1,max=100"` // 每页大小，默认10，最大100
}
