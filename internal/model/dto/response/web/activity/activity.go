package activity

// LeaderInfo 领队简要信息
type LeaderInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// CreateActivityResponse 创建活动响应
type CreateActivityResponse struct {
	ID              string      `json:"id"`
	Leader          *LeaderInfo `json:"leader"`
	Title           string      `json:"title"`
	CoverImg        string      `json:"coverImg"`
	Category        string      `json:"category"`
	Address         string      `json:"address"`
	ActivityContent string      `json:"activityContent"`
	StartTime       string      `json:"startTime"`
	EndTime         string      `json:"endTime"`
	SignUpStart     string      `json:"signUpStart"`
	SignUpEnd       string      `json:"signUpEnd"`
	MaxPeople       int         `json:"maxPeople"`
	CurrentPeople   int         `json:"currentPeople"`
	Price           float64     `json:"price"`
	Difficulty      string      `json:"difficulty"`
	Joined          bool        `json:"joined"`
}
