package home

import responseHome "tuyou/internal/model/dto/response/web/home"

// HomeService 前台首页服务接口
type HomeService interface {
	// GetSlides 获取首页轮播（固定 5 张，不足时按 活动→笔记→攻略 补齐）
	GetSlides() ([]responseHome.SlideItem, error)
}
