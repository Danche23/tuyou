package comment

// 后台评论管理模块 —— 请求 DTO（管理端 → 后端的入参契约）
//
// 注意这里和前台 web 的字段风格不一样：
//   - 前台用驼峰 JSON（targetType、replyToId），后台用 snake_case form 参数
//     （target_type、begin_time）——两套是给不同前端用的。
//   - 后台列表是"多条件筛选"，所以筛选字段全用 form:"xxx" 走 query string，
//     且都标了"空=不筛"——前端不传就不过滤这一项。
//   - Page/Size 用 binding:"required" 强制必传（后台分页必须明确），而前台 reply 列表允许省略。
//
// 中文枚举约定（来自 postman 文档，service 层负责翻译）：
//   target_type：笔记/攻略 → 内部 1/2
//   level：一级/二级 → SQL reply_to_id IS NULL / IS NOT NULL
//   sort_order：升序/降序 → ASC/DESC
//   sort_field：likes/heat → like_count/heat（白名单，防 SQL 注入）

// ListCommentReq 后台评论分页查询请求
type ListCommentReq struct {
	Page       int    `form:"page" binding:"required,min=1"`         // 页码，从1开始
	Size       int    `form:"size" binding:"required,min=1,max=100"` // 每页条数
	TargetType string `form:"target_type"`                           // 目标类型：笔记/攻略，空=不筛
	Nickname   string `form:"nickname"`                              // 评论用户昵称模糊，空=不筛
	Title      string `form:"title"`                                 // 目标标题模糊，空=不筛
	BeginTime  string `form:"begin_time"`                            // 评论创建开始时间 yyyy-MM-dd
	FinishTime string `form:"finish_time"`                           // 评论创建结束时间 yyyy-MM-dd
	Level      string `form:"level"`                                 // 评论层级：一级/二级，空=不筛
	SortField  string `form:"sort_field"`                            // 排序字段：likes/heat，空=默认 id 升序
	SortOrder  string `form:"sort_order"`                            // 排序方式：升序/降序，默认升序
}

// BatchDeleteReq 批量删除评论请求
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1"` // 评论 ID 列表（纯数字）
}
