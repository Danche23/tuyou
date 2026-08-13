package home

import (
	requestHome "tuyou/internal/model/dto/request/web/home"
	responseHome "tuyou/internal/model/dto/response/web/home"
)

// HomeService 前台首页服务接口
type HomeService interface {
	// GetSlides 获取首页轮播（固定 5 张，不足时按 活动→笔记→攻略 补齐）
	GetSlides() ([]responseHome.SlideItem, error)

	// GetFeed 获取Feed流（笔记/攻略跨表合并，瀑布流分页）
	GetFeed(memberID uint, req *requestHome.FeedRequest) (*responseHome.FeedResponse, error)

	// BuildFeedItems 组装 FeedItem 列表（GetFeed 与 search 模块共用的唯一实现）
	BuildFeedItems(memberID uint, entries []FeedEntry) ([]responseHome.FeedItem, error)
}
