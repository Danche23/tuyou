package activity

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
}
