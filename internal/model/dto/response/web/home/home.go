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
