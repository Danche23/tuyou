package home

import "tuyou/pkg/response"

// FeedRequest 获取Feed流请求参数
type FeedRequest struct {
	response.PageRequest
	Category int    `form:"category"` // 分类ID：0=全部，1-7=指定分类
	Sort     string `form:"sort"`     // 排序：推荐/最新/最热，默认推荐
}
