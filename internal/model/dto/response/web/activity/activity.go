package activity

// LeaderInfo 领队简要信息
type LeaderInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ActivityListItem 活动列表项
type ActivityListItem struct {
	ID             string      `json:"id"`
	Leader         *LeaderInfo `json:"leader"`
	Title          string      `json:"title"`
	CoverImg       string      `json:"coverImg"`
	Category       string      `json:"category"`
	Address        string      `json:"address"`
	StartTime      string      `json:"startTime"`
	EndTime        string      `json:"endTime"`
	SignUpEnd      string      `json:"signUpEnd"`
	Price          float64     `json:"price"`
	MaxPeople      int         `json:"maxPeople"`
	CurrentPeople  int         `json:"currentPeople"`
	Difficulty     string      `json:"difficulty"`
	ActivityStatus string      `json:"activityStatus"`
	Joined         bool        `json:"joined"`
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

// MemberInfo 成员简要信息
type MemberInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ActivityDetailResponse 活动详情响应
type ActivityDetailResponse struct {
	ID              string       `json:"id"`
	Leader          *LeaderInfo  `json:"leader"`
	IsOwner         bool         `json:"isOwner"`
	Title           string       `json:"title"`
	CoverImg        string       `json:"coverImg"`
	Category        string       `json:"category"`
	Address         string       `json:"address"`
	ActivityContent string       `json:"activityContent"`
	StartTime       string       `json:"startTime"`
	EndTime         string       `json:"endTime"`
	SignUpStart     string       `json:"signUpStart"`
	SignUpEnd       string       `json:"signUpEnd"`
	MaxPeople       int          `json:"maxPeople"`
	CurrentPeople   int          `json:"currentPeople"`
	Price           float64      `json:"price"`
	Difficulty      string       `json:"difficulty"`
	ActivityStatus  string       `json:"activityStatus"`
	Joined          bool         `json:"joined"`
	Members         []MemberInfo `json:"members"`
}

// SignupToggleResponse 报名/取消报名响应
type SignupToggleResponse struct {
	IsJoined      bool `json:"isJoined"`
	CurrentPeople int  `json:"currentPeople"`
}
