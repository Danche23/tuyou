package guide

// CreateGuideRequest 发布攻略请求
type CreateGuideRequest struct {
	Title      string   `json:"title" binding:"required,max=50"`
	Content    string   `json:"content" binding:"required,max=2000"`
	Category   int      `json:"category" binding:"required,min=1,max=7"`
	Tags       []string `json:"tags" binding:"required,min=1,max=3"`
	Images     []string `json:"images" binding:"required,min=1,max=5"`
	VideoURL   string   `json:"videoUrl" binding:"omitempty,max=255"`
	Distance   string   `json:"distance" binding:"required,max=20"`
	Duration   string   `json:"duration" binding:"required,max=20"`
	Difficulty string   `json:"difficulty" binding:"required"`
	Transport  string   `json:"transport" binding:"required,max=20"`
}
