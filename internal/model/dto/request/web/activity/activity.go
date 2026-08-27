package activity

// ListActivityRequest 活动列表请求
type ListActivityRequest struct {
	Category string `form:"category"`                                 // 可选，活动分类中文名
	Date     string `form:"date"`                                     // 可选，筛选日期（2006-01-02），返回该日期内进行的活动
	Page     int    `form:"page"    binding:"required,min=1"`         // 页码，从1开始
	Size     int    `form:"size"    binding:"required,min=1,max=100"` // 每页条数
}

// CreateActivityRequest 创建活动请求
type CreateActivityRequest struct {
	Title           string  `json:"title"           binding:"required,max=50"`
	Category        string  `json:"category"        binding:"required"`
	CoverImg        string  `json:"coverImg"        binding:"required,max=255"`
	Address         string  `json:"address"         binding:"required,max=200"`
	StartTime       string  `json:"startTime"       binding:"required"`
	EndTime         string  `json:"endTime"         binding:"required"`
	Price           float64 `json:"price"           binding:"gte=0"`
	MaxPeople       int     `json:"maxPeople"       binding:"required,min=1"`
	Difficulty      string  `json:"difficulty"      binding:"required"`
	ActivityContent string  `json:"activityContent" binding:"required,max=2000"`
	WechatQR        string  `json:"wechatQR"        binding:"max=500"`
}
