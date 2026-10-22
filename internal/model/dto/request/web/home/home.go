package home

// FeedRequest 获取Feed流请求参数
//
// 分页支持两种模式：
//   - 游标分页（推荐）：首次不带 cursor，取下一页时把上次响应里的 next_cursor 原样带回
//   - 页码分页（兼容）：带 page/size，行为与历史一致
//
// 注意这里不内嵌 response.PageRequest：它的 page/size 带 binding:"required"，
// 会让不带 page 的游标请求被 ShouldBindQuery 直接拒成 400；而 PageRequest 为多个接口共用，
// 不能为了本接口放宽它。改为自行声明后，缺省值由 service 层兜底（page<1→1、size<1→10、size>100→100）。
type FeedRequest struct {
	Page     int    `form:"page"`     // 页码（兼容模式），从1开始，缺省1
	Size     int    `form:"size"`     // 每页条数，缺省10，上限100
	Category int    `form:"category"` // 分类ID：0=全部，1-7=指定分类
	Sort     string `form:"sort"`     // 排序：推荐/最新/最热，默认推荐
	Cursor   string `form:"cursor"`   // 分页游标；有值时走游标分页，此时忽略 page
}
